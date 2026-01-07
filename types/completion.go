package types

type Completion struct {
	Messages []Message `json:"messages" binding:"required,min=1,max=100,dive"`
	Model    string    `json:"model" binding:"omitempty,min=1,max=100"`
	Stream   bool      `json:"stream"`
	Tier     string    `json:"tier,omitempty" binding:"omitempty,oneof=budget standard premium ultra-premium"`
	// Optional fields
	Temperature      *float64 `json:"temperature,omitempty" binding:"omitempty,gte=0,lte=2"`
	MaxTokens        *int     `json:"max_tokens,omitempty" binding:"omitempty,gt=0,lte=100000"`
	TopP             *float64 `json:"top_p,omitempty" binding:"omitempty,gte=0,lte=1"`
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty" binding:"omitempty,gte=-2,lte=2"`
	PresencePenalty  *float64 `json:"presence_penalty,omitempty" binding:"omitempty,gte=-2,lte=2"`
}
