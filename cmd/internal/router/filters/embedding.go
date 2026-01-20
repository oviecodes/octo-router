package filters

import (
	"context"
	"fmt"
	"llm-router/cmd/internal/providers"
	"llm-router/types"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sugarme/tokenizer"
	"github.com/sugarme/tokenizer/pretrained"
	ort "github.com/yalue/onnxruntime_go"
)

const (
	MaxSeqLength = 128
	EmbeddingDim = 384
)

type EmbeddingFilter struct {
	policy           *types.SemanticPolicy
	tokenizer        *tokenizer.Tokenizer
	session          *ort.AdvancedSession
	intentEmbeddings map[string][]float32
	mu               sync.Mutex

	// Pre-allocated buffers for inference
	inputIds      []int64
	attentionMask []int64
	tokenTypeIds  []int64
	outputData    []float32
}

func NewEmbeddingFilter(policy *types.SemanticPolicy) (*EmbeddingFilter, error) {
	// 1. Initialize ONNX runtime if not already done
	if !ort.IsInitialized() {
		// Set shared library path if needed (e.g. for macOS Homebrew)
		if _, err := os.Stat("/opt/homebrew/lib/libonnxruntime.dylib"); err == nil {
			ort.SetSharedLibraryPath("/opt/homebrew/lib/libonnxruntime.dylib")
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
	}

	err := f.loadModel(policy.ModelPath)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (f *EmbeddingFilter) Name() string {
	return "embedding"
}

func (f *EmbeddingFilter) loadModel(path string) error {
	// Try to find the file if the path is relative and doesn't exist directly
	finalPath := path
	if _, err := os.Stat(finalPath); os.IsNotExist(err) && !filepath.IsAbs(finalPath) {
		// Try project root from cmd/server or cmd/internal...
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

	// 1. Load Tokenizer
	tk := tokenizer.NewTokenizerFromFile(filepath.Join(filepath.Dir(finalPath), "tokenizer.json"))
	if tk == nil {
		// Fallback to Bert if custom tokenizer missing
		f.tokenizer = pretrained.BertBaseUncased()
	} else {
		f.tokenizer = tk
	}

	// 2. Create Tensors
	shape := ort.NewShape(1, MaxSeqLength)
	inputIdsTensor, err := ort.NewTensor(shape, f.inputIds)
	if err != nil {
		return err
	}
	maskTensor, err := ort.NewTensor(shape, f.attentionMask)
	if err != nil {
		return err
	}
	typeIdsTensor, err := ort.NewTensor(shape, f.tokenTypeIds)
	if err != nil {
		return err
	}
	outputShape := ort.NewShape(1, MaxSeqLength, EmbeddingDim)
	outputTensor, err := ort.NewTensor(outputShape, f.outputData)
	if err != nil {
		return err
	}

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

	// 4. Pre-calculate Intent Embeddings (Few-Shot Centroids)
	for _, group := range f.policy.Groups {
		texts := make([]string, 0)
		if group.IntentDescription != "" {
			texts = append(texts, group.IntentDescription)
		}
		texts = append(texts, group.Examples...)

		if len(texts) == 0 && len(group.IntentKeywords) > 0 {
			texts = append(texts, strings.Join(group.IntentKeywords, " "))
		}

		if len(texts) > 0 {
			centroid := make([]float32, EmbeddingDim)
			for _, t := range texts {
				emb, err := f.calculateEmbedding(context.Background(), t)
				if err != nil {
					return fmt.Errorf("failed to calculate embedding for text %q in group %s: %w", t, group.Name, err)
				}
				for d := 0; d < EmbeddingDim; d++ {
					centroid[d] += emb[d]
				}
			}
			// Average
			for d := 0; d < EmbeddingDim; d++ {
				centroid[d] /= float32(len(texts))
			}
			f.intentEmbeddings[group.Name] = centroid
		}
	}

	return nil
}

func (f *EmbeddingFilter) calculateEmbedding(ctx context.Context, text string) ([]float32, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// 1. Tokenize
	en, err := f.tokenizer.EncodeSingle(text, true)
	if err != nil {
		return nil, err
	}

	ids := en.GetIds()
	mask := en.GetAttentionMask()
	typeIds := en.GetTypeIds()

	// CRITICAL: Ensure [CLS] (101) and [SEP] (102) are present
	if len(ids) > 0 && ids[0] != 101 {
		newIds := make([]int, 0, len(ids)+2)
		newIds = append(newIds, 101)
		newIds = append(newIds, ids...)
		if newIds[len(newIds)-1] != 102 {
			newIds = append(newIds, 102)
		}
		ids = newIds
		mask = make([]int, len(ids))
		typeIds = make([]int, len(ids))
		for i := range mask {
			mask[i] = 1
		}
	}

	// 2. Prepare Inputs (with padding)
	for i := 0; i < MaxSeqLength; i++ {
		if i < len(ids) {
			f.inputIds[i] = int64(ids[i])
			f.attentionMask[i] = int64(mask[i])
			f.tokenTypeIds[i] = int64(typeIds[i])
		} else {
			f.inputIds[i] = 0
			f.attentionMask[i] = 0
			f.tokenTypeIds[i] = 0
		}
	}

	// 3. Run Inference
	err = f.session.Run()
	if err != nil {
		return nil, err
	}

	// 4. Mean Pooling (Modified: Re-include structural tokens for better alignment)
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

	return embedding, nil
}

func (f *EmbeddingFilter) Filter(ctx context.Context, candidates []types.Provider, input *types.SelectProviderInput) ([]types.Provider, error) {
	if len(input.Messages) == 0 {
		return candidates, nil
	}

	// 1. Get Prompt Embedding
	lastMsg := input.Messages[len(input.Messages)-1].Content
	promptEmb, err := f.calculateEmbedding(ctx, lastMsg)
	if err != nil {
		return candidates, err
	}

	// 2. Find Best Match
	bestGroup := f.policy.DefaultGroup
	maxSim := -1.0

	for name, intentEmb := range f.intentEmbeddings {
		sim := cosineSimilarity(promptEmb, intentEmb)
		fmt.Printf("[Semantic Debug] Group: %s, Score: %.4f\n", name, sim)
		if sim > maxSim {
			maxSim = sim
			bestGroup = name
		}
	}

	threshold := f.policy.Threshold
	if threshold == 0 {
		threshold = 0.5
	}

	if maxSim < threshold {
		fmt.Printf("[Semantic] Max similarity %.4f below threshold %.4f, using default: %s\n", maxSim, threshold, f.policy.DefaultGroup)
		bestGroup = f.policy.DefaultGroup
	} else {
		fmt.Printf("[Semantic] Matched intent: %s (score: %.4f)\n", bestGroup, maxSim)
	}

	// 3. Discovery Logic (Hybrid)
	var allowList []string
	var reqCapability string

	for _, group := range f.policy.Groups {
		if group.Name == bestGroup {
			allowList = group.AllowProviders
			reqCapability = group.RequiredCapability
			break
		}
	}

	if len(allowList) == 0 && reqCapability != "" {
		allowList = providers.ListProvidersByCapability(reqCapability)
	}

	// 4. Apply Filter
	if len(allowList) == 0 {
		return candidates, nil
	}

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
