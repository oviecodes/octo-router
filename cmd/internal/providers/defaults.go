package providers

import "llm-router/types"

func GetDefaultCatalog() []types.ModelConfig {
	return []types.ModelConfig{
		// OpenAI Models
		{
			ID:              "openai/gpt-5.1",
			Provider:        "openai",
			Name:            "GPT-5.1",
			InputCostPer1M:  7.50,
			OutputCostPer1M: 22.50,
			ContextWindow:   200000,
			Tier:            "ultra-premium",
		},
		{
			ID:              "openai/gpt-5",
			Provider:        "openai",
			Name:            "GPT-5",
			InputCostPer1M:  5.00,
			OutputCostPer1M: 15.00,
			ContextWindow:   200000,
			Tier:            "ultra-premium",
		},
		{
			ID:              "openai/gpt-4o",
			Provider:        "openai",
			Name:            "GPT-4o",
			InputCostPer1M:  2.50,
			OutputCostPer1M: 10.00,
			ContextWindow:   128000,
			Tier:            "premium",
		},
		{
			ID:              "openai/gpt-3.5-turbo",
			Provider:        "openai",
			Name:            "GPT-3.5 Turbo",
			InputCostPer1M:  0.50,
			OutputCostPer1M: 1.50,
			ContextWindow:   16385,
			Tier:            "standard",
		},
		{
			ID:              "openai/gpt-4o-mini",
			Provider:        "openai",
			Name:            "GPT-4o Mini",
			InputCostPer1M:  0.15,
			OutputCostPer1M: 0.60,
			ContextWindow:   128000,
			Tier:            "budget",
		},

		// Anthropic Models
		{
			ID:              "anthropic/claude-opus-4.5",
			Provider:        "anthropic",
			Name:            "Claude Opus 4.5",
			InputCostPer1M:  15.00,
			OutputCostPer1M: 75.00,
			ContextWindow:   200000,
			Tier:            "ultra-premium",
		},
		{
			ID:              "anthropic/claude-sonnet-4",
			Provider:        "anthropic",
			Name:            "Claude Sonnet 4",
			InputCostPer1M:  3.00,
			OutputCostPer1M: 15.00,
			ContextWindow:   200000,
			Tier:            "premium",
		},
		{
			ID:              "anthropic/claude-haiku-4.5",
			Provider:        "anthropic",
			Name:            "Claude Haiku 4.5",
			InputCostPer1M:  0.80,
			OutputCostPer1M: 4.00,
			ContextWindow:   200000,
			Tier:            "standard",
		},
		{
			ID:              "anthropic/claude-haiku-3",
			Provider:        "anthropic",
			Name:            "Claude Haiku 3",
			InputCostPer1M:  0.25,
			OutputCostPer1M: 1.25,
			ContextWindow:   200000,
			Tier:            "standard",
		},

		// Gemini Models
		{
			ID:              "gemini/gemini-3-pro",
			Provider:        "gemini",
			Name:            "Gemini 3.0 Pro",
			InputCostPer1M:  2.00,
			OutputCostPer1M: 12.00,
			ContextWindow:   1000000,
			Tier:            "premium",
		},
		{
			ID:              "gemini/gemini-2.5-pro",
			Provider:        "gemini",
			Name:            "Gemini 2.5 Pro",
			InputCostPer1M:  1.25,
			OutputCostPer1M: 10.00,
			ContextWindow:   1000000,
			Tier:            "premium",
		},
		{
			ID:              "gemini/gemini-3-flash",
			Provider:        "gemini",
			Name:            "Gemini 3.0 Flash",
			InputCostPer1M:  0.50,
			OutputCostPer1M: 3.00,
			ContextWindow:   1000000,
			Tier:            "standard",
		},
		{
			ID:              "gemini/gemini-2.5-flash",
			Provider:        "gemini",
			Name:            "Gemini 2.5 Flash",
			InputCostPer1M:  0.30,
			OutputCostPer1M: 2.50,
			ContextWindow:   1000000,
			Tier:            "standard",
		},
		{
			ID:              "gemini/gemini-2.5-flash-lite",
			Provider:        "gemini",
			Name:            "Gemini 2.5 Flash Lite",
			InputCostPer1M:  0.10,
			OutputCostPer1M: 0.40,
			ContextWindow:   1000000,
			Tier:            "budget",
		},
	}
}
