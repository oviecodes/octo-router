package filters

import (
	"context"
	"fmt"
	"llm-router/cmd/internal/providers"
	"llm-router/types"
	"math"
	"os"
	"path/filepath"
	"sync"

	"github.com/sugarme/tokenizer"
	"github.com/sugarme/tokenizer/pretrained"
	ort "github.com/yalue/onnxruntime_go"
	"go.uber.org/zap"
)

// Constants for the MiniLM-L6-v2 model
const (
	MaxSeqLength = 128 // Transformer input limit
	EmbeddingDim = 384 // MiniLM output vector size
)

// EmbeddingFilter implements the ProviderFilter interface using local vector similarity.
// It converts user prompts into mathematical vectors (embeddings) and compares them
// against pre-calculated "intent clusters" to determine the most relevant routing.
type EmbeddingFilter struct {
	policy           *types.SemanticPolicy
	tokenizer        *tokenizer.Tokenizer
	session          *ort.AdvancedSession
	intentEmbeddings map[string][]float32
	mu               sync.Mutex
	logger           *zap.Logger

	// Pre-allocated buffers for inference (optimized for performance)
	inputIds      []int64
	attentionMask []int64
	tokenTypeIds  []int64
	outputData    []float32
}

// NewEmbeddingFilter creates and initializes a new EmbeddingFilter.
// It handles ONNX Runtime environment setup and model loading.
func NewEmbeddingFilter(policy *types.SemanticPolicy, logger *zap.Logger) (*EmbeddingFilter, error) {
	// 1. Initialize ONNX runtime environment
	if !ort.IsInitialized() {
		libPath := resolveLibPath(policy.SharedLibPath, logger)
		if libPath != "" {
			ort.SetSharedLibraryPath(libPath)
		}

		err := ort.InitializeEnvironment()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize onnxruntime: %w", err)
		}
	}

	f := &EmbeddingFilter{
		policy:           policy,
		intentEmbeddings: make(map[string][]float32),
		inputIds:         make([]int64, MaxSeqLength),
		attentionMask:    make([]int64, MaxSeqLength),
		tokenTypeIds:     make([]int64, MaxSeqLength),
		outputData:       make([]float32, MaxSeqLength*EmbeddingDim),
		logger:           logger,
	}

	err := f.loadModel(policy.ModelPath)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// resolveLibPath finds the best ONNX shared library path based on priority:
// 1. Environment variable (ONNXRUNTIME_LIB_PATH)
// 2. Configuration file (shared_lib_path)
// 3. Known system locations (e.g. Homebrew)
func resolveLibPath(configPath string, logger *zap.Logger) string {
	// 1. Try Env Var
	if env := os.Getenv("ONNXRUNTIME_LIB_PATH"); env != "" {
		if logger != nil {
			logger.Debug("Using ONNX library from environment variable", zap.String("path", env))
		}
		return env
	}

	// 2. Try Config
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			if logger != nil {
				logger.Debug("Using ONNX library from config", zap.String("path", configPath))
			}
			return configPath
		}
	}

	// 3. Known System Paths (macOS/Linux)
	defaults := []string{
		"/usr/local/lib/libonnxruntime.dylib",
		"/opt/homebrew/lib/libonnxruntime.dylib",
		"/usr/lib/libonnxruntime.so",
	}
	for _, p := range defaults {
		if _, err := os.Stat(p); err == nil {
			if logger != nil {
				logger.Debug("Using ONNX library from default path", zap.String("path", p))
			}
			return p
		}
	}

	return ""
}

func (f *EmbeddingFilter) Name() string {
	return "embedding"
}

// loadModel prepares the tokenizer and ONNX inference session.
func (f *EmbeddingFilter) loadModel(path string) error {
	// Resolve relative model path
	finalPath := path
	if _, err := os.Stat(finalPath); os.IsNotExist(err) && !filepath.IsAbs(finalPath) {
		tryPaths := []string{
			filepath.Join("..", "..", path),
			filepath.Join("..", path),
		}
		for _, tp := range tryPaths {
			if _, err := os.Stat(tp); err == nil {
				finalPath = tp
				break
			}
		}
	}

	// 1. Load Tokenizer (custom tokenizer.json if present, else fallback)
	tk := tokenizer.NewTokenizerFromFile(filepath.Join(filepath.Dir(finalPath), "tokenizer.json"))
	if tk == nil {
		f.tokenizer = pretrained.BertBaseUncased()
	} else {
		f.tokenizer = tk
	}

	// 2. Bind Tensors to Buffers
	shape := ort.NewShape(1, MaxSeqLength)
	inputIdsTensor, _ := ort.NewTensor(shape, f.inputIds)
	maskTensor, _ := ort.NewTensor(shape, f.attentionMask)
	typeIdsTensor, _ := ort.NewTensor(shape, f.tokenTypeIds)
	outputShape := ort.NewShape(1, MaxSeqLength, EmbeddingDim)
	outputTensor, _ := ort.NewTensor(outputShape, f.outputData)

	// 3. Create Session
	session, err := ort.NewAdvancedSession(finalPath,
		[]string{"input_ids", "attention_mask", "token_type_ids"},
		[]string{"last_hidden_state"},
		[]ort.Value{inputIdsTensor, maskTensor, typeIdsTensor},
		[]ort.Value{outputTensor},
		nil)
	if err != nil {
		return fmt.Errorf("failed to create onnx session: %w", err)
	}
	f.session = session

	// 4. Pre-calculate Intent Centroids (Averaged Few-Shot Vectors)
	f.initializeIntentClusters()

	return nil
}

func (f *EmbeddingFilter) initializeIntentClusters() {
	// Merge user groups with system defaults if enabled
	allGroups := f.policy.Groups
	if f.policy.ExtendDefault {
		defaults := GetSystemDefaultGroups()
		// For each default group, either merge with user group or add new
		for _, dg := range defaults {
			found := false
			for i, ug := range allGroups {
				if ug.Name == dg.Name && ug.UseSystemDefault {
					allGroups[i].Examples = append(allGroups[i].Examples, dg.Examples...)
					found = true
					break
				}
			}
			if !found {
				// Don't add default groups unless specifically relevant or user didn't provide alternatives
				// For now, let's keep it simple: just merge if the name matches
			}
		}
	}

	for _, group := range allGroups {
		texts := make([]string, 0)
		if group.IntentDescription != "" {
			texts = append(texts, group.IntentDescription)
		}
		texts = append(texts, group.Examples...)

		if len(texts) > 0 {
			centroid := make([]float32, EmbeddingDim)
			validSampleCount := 0
			for _, t := range texts {
				emb, err := f.calculateEmbedding(context.Background(), t)
				if err != nil {
					continue
				}
				for d := 0; d < EmbeddingDim; d++ {
					centroid[d] += emb[d]
				}
				validSampleCount++
			}
			// Average to get the cluster center
			if validSampleCount > 0 {
				for d := 0; d < EmbeddingDim; d++ {
					centroid[d] /= float32(validSampleCount)
				}
				f.intentEmbeddings[group.Name] = centroid
				f.logger.Info("Initialized intent cluster", zap.String("group", group.Name), zap.Int("samples", validSampleCount))
			}
		}
	}
}

// calculateEmbedding processes a string through the neural network.
func (f *EmbeddingFilter) calculateEmbedding(ctx context.Context, text string) ([]float32, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// 1. Tokenization: Break text into sub-words (WordPiece)
	en, err := f.tokenizer.EncodeSingle(text, true)
	if err != nil {
		return nil, err
	}

	ids := en.GetIds()
	mask := en.GetAttentionMask()
	typeIds := en.GetTypeIds()

	// Ensure BERT special tokens are present ([CLS] at 0, [SEP] at end)
	if len(ids) > 0 && ids[0] != 101 {
		newIds := []int{101}
		newIds = append(newIds, ids...)
		if newIds[len(newIds)-1] != 102 {
			newIds = append(newIds, 102)
		}
		ids = newIds
		// Resync mask
		mask = make([]int, len(ids))
		for i := range mask {
			mask[i] = 1
		}
		typeIds = make([]int, len(ids))
	}

	// 2. Data Preparation: Fill ONNX buffers with padding
	for i := 0; i < MaxSeqLength; i++ {
		if i < len(ids) {
			f.inputIds[i] = int64(ids[i])
			f.attentionMask[i] = int64(mask[i])
			f.tokenTypeIds[i] = int64(typeIds[i])
		} else {
			f.inputIds[i] = 0 // [PAD] id
			f.attentionMask[i] = 0
			f.tokenTypeIds[i] = 0
		}
	}

	// 3. Inference: Run the Transformer model
	err = f.session.Run()
	if err != nil {
		return nil, err
	}

	// 4. Mean Pooling: Average the vectors of all tokens in the sentence
	// This reduces a matrix (Words x Dimensions) into a single vector (Dimensions).
	embedding := make([]float32, EmbeddingDim)
	var validTokens float32

	for i := 0; i < MaxSeqLength; i++ {
		if i < len(ids) {
			validTokens++
			for d := 0; d < EmbeddingDim; d++ {
				embedding[d] += f.outputData[i*EmbeddingDim+d]
			}
		}
	}

	if validTokens > 0 {
		for d := 0; d < EmbeddingDim; d++ {
			embedding[d] /= validTokens
		}
	}

	// Calculate magnitude for debug
	var mag float64
	for _, v := range embedding {
		mag += float64(v) * float64(v)
	}
	magnitude := math.Sqrt(mag)

	f.logger.Debug("Point embedding calculated",
		zap.String("text", text),
		zap.Int("tokens", len(ids)),
		zap.Float64("magnitude", magnitude),
	)

	return embedding, nil
}

// Filter compares the prompt's vector against each intent's centroid vector.
func (f *EmbeddingFilter) Filter(ctx context.Context, input *types.FilterInput) (*types.FilterOutput, error) {
	candidates := input.Candidates
	if len(input.Messages) == 0 {
		return &types.FilterOutput{Candidates: candidates}, nil
	}

	// 1. Vectorize the prompt
	lastMsg := input.Messages[len(input.Messages)-1].Content
	promptEmb, err := f.calculateEmbedding(ctx, lastMsg)
	if err != nil {
		return &types.FilterOutput{Candidates: candidates}, err
	}

	// 2. Find closest intent cluster (Cosine Similarity)
	bestGroup := f.policy.DefaultGroup
	maxSim := -1.0

	for name, intentEmb := range f.intentEmbeddings {
		sim := cosineSimilarity(promptEmb, intentEmb)
		f.logger.Debug("Semantic Group Score",
			zap.String("group", name),
			zap.Float64("score", sim),
		)
		if sim > maxSim {
			maxSim = sim
			bestGroup = name
		}
	}

	// 3. Apply Decision Threshold
	threshold := f.policy.Threshold
	if threshold == 0 {
		threshold = 0.5
	}

	if maxSim < threshold {
		f.logger.Info("Semantic similarity below threshold, using default group",
			zap.Float64("max_sim", maxSim),
			zap.Float64("threshold", threshold),
			zap.String("default_group", f.policy.DefaultGroup),
		)
		bestGroup = f.policy.DefaultGroup
	} else {
		f.logger.Info("Semantic match found",
			zap.String("intent", bestGroup),
			zap.Float64("score", maxSim),
		)
	}

	// 4. Transform Decision into Candidate Pool
	filtered, err := f.resolveCandidates(bestGroup, candidates)
	return &types.FilterOutput{Candidates: filtered}, err
}

func (f *EmbeddingFilter) resolveCandidates(groupName string, candidates []types.Provider) ([]types.Provider, error) {
	var allowList []string
	var reqCapability string

	// Find the matching group configuration
	for _, group := range f.policy.Groups {
		if group.Name == groupName {
			allowList = group.AllowProviders
			reqCapability = group.RequiredCapability
			break
		}
	}

	// If no group config found but we have defaults, try system defaults
	if len(allowList) == 0 && reqCapability == "" && f.policy.ExtendDefault {
		for _, dg := range GetSystemDefaultGroups() {
			if dg.Name == groupName {
				allowList = dg.AllowProviders
				reqCapability = dg.RequiredCapability
				break
			}
		}
	}

	// Capability discovery if no explicit provider list
	if len(allowList) == 0 && reqCapability != "" {
		allowList = providers.ListProvidersByCapability(reqCapability)
	}

	if len(allowList) == 0 {
		return candidates, nil
	}

	// Final Filtering
	filtered := make([]types.Provider, 0)
	allowedMap := make(map[string]bool)
	for _, p := range allowList {
		allowedMap[p] = true
	}

	for _, p := range candidates {
		if allowedMap[p.GetProviderName()] {
			filtered = append(filtered, p)
		}
	}

	if len(filtered) == 0 {
		return candidates, nil
	}

	return filtered, nil
}

// cosineSimilarity measures how "aligned" two vectors are.
// Scores range from -1.0 (opposite) to 1.0 (identical).
func cosineSimilarity(a, b []float32) float64 {
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
