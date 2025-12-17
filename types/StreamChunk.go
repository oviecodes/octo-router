package types

type StreamChunk struct {
	Content string
	Done    bool
	Error   error
}
