package types

type Message struct {
	Role    string `json:"role" binding:"required,oneof=user assistant system"`
	Content string `json:"content" binding:"required,min=1,max=500000"`
}
