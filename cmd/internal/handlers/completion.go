package handlers

import (
	"fmt"
	"llm-router/cmd/internal/app"
	"llm-router/cmd/internal/providers"
	"llm-router/cmd/internal/validations"
	"llm-router/types"
	"llm-router/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var logger = utils.SetUpLogger()

func HandleStreamingCompletion(resolver app.ConfigResolver, c *gin.Context, provider providers.Provider, request types.Completion) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	chunks, err := provider.CompleteStream(c.Request.Context(), request.Messages)
	if err != nil {
		logger.Error("Provider streaming failed", zap.Error(err))
		c.SSEvent("error", gin.H{
			"error": "Failed to start streaming completion",
		})
		return
	}

	// Stream chunks to client
	for chunk := range chunks {
		c.SSEvent("message", chunk)
		c.Writer.Flush()

		if chunk.Done {
			break
		}
	}
}

func Completions(resolver app.ConfigResolver, c *gin.Context) {
	var request types.Completion

	if err := c.ShouldBindJSON(&request); err != nil {
		validations.HandleValidationError(c, err)
		return
	}

	if err := validateCompletionRequest(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	logger.Info("Completion request received",
		zap.Int("message_count", len(request.Messages)),
		zap.String("model", request.Model),
		zap.Bool("stream", request.Stream),
	)

	provider := resolver.GetRouter(c).SelectProvider(c.Request.Context())

	if request.Stream {
		HandleStreamingCompletion(resolver, c, provider, request)
	} else {
		// TODO: Call provider and return response - move to handler file
		response, err := provider.Complete(c.Request.Context(), request.Messages)
		if err != nil {
			logger.Error("Provider completion failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate completion",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  response.Content,
			"role":     response.Role,
			"provider": fmt.Sprintf("%T", provider),
		})
	}

}

// validateCompletionRequest performs additional business logic validation
func validateCompletionRequest(req *types.Completion) error {

	if len(req.Messages) > 0 {

		if req.Messages[0].Role != "user" && req.Messages[0].Role != "system" {
			return fmt.Errorf("first message must be from user or system")
		}
	}

	if req.Temperature != nil {
		if *req.Temperature < 0 || *req.Temperature > 2 {
			return fmt.Errorf("temperature must be between 0 and 2")
		}
	}

	// Check total message length (approximate)
	totalLength := 0
	for _, msg := range req.Messages {
		totalLength += len(msg.Content)
	}
	if totalLength > 1000000 { // 1MB limit
		return fmt.Errorf("total message content too large (max 1MB)")
	}

	return nil
}
