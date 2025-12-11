package types

type ModelData struct {
	DefaultModels map[string]DefaultModels `mapstructure:"defaults"`
}

type DefaultModels struct {
	// Name      string `mapstructure:"name"`
	Model     string `mapstructure:"model"`
	MaxTokens int64  `mapstructure:"maxTokens"`
}

type ResilienceData struct {
	Timeout              int            `mapstructure:"timeout"`
	RetriesConfig        map[string]int `mapstructure:"retries"`
	CircuitBreakerConfig map[string]int `mapstructure:"circuitBreaker"`
}

type LimitsData struct {
	RequestsPerMinute int                       `mapstructure:"requestsPerMinute"`
	RequestsPerDay    int                       `mapstructure:"requestsPerDay"`
	DailyBudget       float64                   `mapstructure:"dailyBudget"`
	AlertThreshold    float64                   `mapstructure:"alertThreshold"`
	Providers         map[string]ProviderLimits `mapstructure:"providers"`
}

type ProviderLimits struct {
	RequestsPerMinute int `mapstructure:"requestsPerMinute"`
}
