package filters

import (
	"context"
	"llm-router/types"
	"strings"

	"go.uber.org/zap"
)

type KeywordFilter struct {
	policy *types.SemanticPolicy
	logger *zap.Logger
}

func NewKeywordFilter(policy *types.SemanticPolicy, logger *zap.Logger) *KeywordFilter {
	return &KeywordFilter{
		policy: policy,
		logger: logger,
	}
}

func (f *KeywordFilter) Name() string {
	return "Semantic(Keyword)"
}

func (f *KeywordFilter) Filter(ctx context.Context, input *types.FilterInput) (*types.FilterOutput, error) {
	candidates := input.Candidates
	if f.policy == nil || !f.policy.Enabled {
		return &types.FilterOutput{Candidates: candidates}, nil
	}

	var promptBuilder strings.Builder
	for _, msg := range input.Messages {
		promptBuilder.WriteString(msg.Content)
		promptBuilder.WriteString(" ")
	}
	prompt := strings.ToLower(promptBuilder.String())

	matchedGroup := f.policy.DefaultGroup

	for _, group := range f.policy.Groups {
		for _, keyword := range group.IntentKeywords {
			if strings.Contains(prompt, strings.ToLower(keyword)) {
				matchedGroup = group.Name
				f.logger.Debug("Keyword match found", zap.String("keyword", keyword), zap.String("group", group.Name))
				break
			}
		}
		if matchedGroup != f.policy.DefaultGroup {
			break
		}
	}

	if matchedGroup == f.policy.DefaultGroup {
		f.logger.Info("No keyword match found, using default group", zap.String("default_group", matchedGroup))
	} else {
		f.logger.Info("Semantic match found (Keyword)", zap.String("intent", matchedGroup))
	}

	var allowList []string
	foundGroup := false

	for _, group := range f.policy.Groups {
		if group.Name == matchedGroup {
			allowList = group.AllowProviders
			foundGroup = true
			break
		}
	}

	if !foundGroup {
		return &types.FilterOutput{Candidates: candidates}, nil
	}

	if len(allowList) == 0 {
		return &types.FilterOutput{Candidates: candidates}, nil
	}

	var filtered []types.Provider
	for _, p := range candidates {
		name := p.GetProviderName()
		allowed := false
		for _, allowedName := range allowList {
			if strings.EqualFold(name, allowedName) {
				allowed = true
				break
			}
		}
		if allowed {
			filtered = append(filtered, p)
		}
	}

	return &types.FilterOutput{Candidates: filtered}, nil
}
