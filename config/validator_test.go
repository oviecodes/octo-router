package config

import (
	"llm-router/types"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "Valid Config",
			config: Config{
				Providers: []types.ProviderConfig{
					{Name: "openai", Enabled: true},
				},
				Resilience: types.ResilienceData{Timeout: 30},
				Routing: types.RoutingData{
					Strategy: "round-robin",
				},
			},
			wantErr: false,
		},
		{
			name: "No Enabled Providers",
			config: Config{
				Providers: []types.ProviderConfig{
					{Name: "openai", Enabled: false},
				},
				Resilience: types.ResilienceData{Timeout: 30},
			},
			wantErr: true,
		},
		{
			name: "Valid Weighted Sum (Not 100)",
			config: Config{
				Providers: []types.ProviderConfig{
					{Name: "openai", Enabled: true},
					{Name: "anthropic", Enabled: true},
				},
				Routing: types.RoutingData{
					Strategy: "weighted",
					Weights: map[string]int{
						"openai":    50,
						"anthropic": 40, // Sum = 90 (now valid due to normalization)
					},
				},
				Resilience: types.ResilienceData{Timeout: 30},
			},
			wantErr: false,
		},
		{
			name: "Negative Weight",
			config: Config{
				Providers: []types.ProviderConfig{
					{Name: "openai", Enabled: true},
				},
				Routing: types.RoutingData{
					Strategy: "weighted",
					Weights: map[string]int{
						"openai": -10,
					},
				},
				Resilience: types.ResilienceData{Timeout: 30},
			},
			wantErr: true,
		},
		{
			name: "Zero Timeout",
			config: Config{
				Providers: []types.ProviderConfig{
					{Name: "openai", Enabled: true},
				},
				Resilience: types.ResilienceData{Timeout: 0},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
