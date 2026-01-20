package filters

import (
	"context"
	"llm-router/types"
	"strings"
)

type KeywordFilter struct {
	policy *types.SemanticPolicy
}

func NewKeywordFilter(policy *types.SemanticPolicy) *KeywordFilter {
	return &KeywordFilter{
		policy: policy,
	}
}

func (f *KeywordFilter) Name() string {
	return "Semantic(Keyword)"
}

func (f *KeywordFilter) Filter(ctx context.Context, candidates []types.Provider, input *types.SelectProviderInput) ([]types.Provider, error) {
	if f.policy == nil || !f.policy.Enabled {
		return candidates, nil
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
				break
			}
		}
		if matchedGroup != f.policy.DefaultGroup {
			break
		}
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
		return candidates, nil
	}

	if len(allowList) == 0 {
		return candidates, nil
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

	return filtered, nil
}
