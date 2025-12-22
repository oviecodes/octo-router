package handlers

import (
	"context"
	"fmt"
	"llm-router/cmd/internal/app"
	"llm-router/cmd/internal/resilience"
	"llm-router/cmd/internal/validations"
	"llm-router/types"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func HandleStreamingCompletion(resolver app.ConfigResolver, c *gin.Context, provider types.Provider, request types.Completion) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	chunks, err := provider.CompleteStream(c.Request.Context(), request.Messages)

	fmt.Printf("Streaming error %v \n", err)
	if err != nil {
		resolver.GetLogger(c).Error("Provider streaming failed", zap.Error(err))
		c.SSEvent("error", gin.H{
			"error": "Failed to start streaming completion",
		})
		return
	}

	// Stream chunks to client
	for chunk := range chunks {
		if chunk.Error != nil {
			c.SSEvent("error", gin.H{
				"error": chunk.Error.Error(),
			})
			c.Writer.Flush()
			break
		}

		c.SSEvent("message", chunk)
		c.Writer.Flush()

		if chunk.Done {
			break
		}
	}
}

func Completions(resolver app.ConfigResolver, c *gin.Context) {
	var request types.Completion
	retry := resolver.GetRetry(c)
	circuitBreakers := resolver.GetCircuitBreaker(c)

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

	resolver.GetLogger(c).Info("Completion request received",
		zap.Int("message_count", len(request.Messages)),
		zap.String("model", request.Model),
		zap.Bool("stream", request.Stream),
	)

	router := resolver.GetRouter(c)
	provider, err := router.SelectProvider(c.Request.Context(), circuitBreakers)

	if err != nil {
		err := fmt.Errorf("no avialable providers, cannot process requests")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	providerName := provider.GetProviderName()
	circuitBreaker := circuitBreakers[providerName]

	fmt.Printf("current circuit breaker failure count: %v, state: %v \n", circuitBreaker.GetState(), circuitBreaker)

	if request.Stream {
		HandleStreamingCompletion(resolver, c, provider, request)
	} else {

		response, err := resilience.Do(c, retry, func(ctx context.Context) (*types.Message, error) {
			return provider.Complete(c.Request.Context(), request.Messages)
		})

		fmt.Printf("error from %T provider: %v", provider, err)

		circuitBreaker.Execute(err)

		if err != nil {
			resolver.GetLogger(c).Error("Provider completion failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
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
