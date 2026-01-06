package handlers

import (
	"context"
	"fmt"
	"llm-router/cmd/internal/app"
	"llm-router/cmd/internal/providers"
	"llm-router/cmd/internal/resilience"
	"llm-router/cmd/internal/validations"
	"llm-router/types"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func buildProviderChain(primaryProvider types.Provider, fallbackNames []string, manager *providers.ProviderManager) []types.Provider {
	providerChain := make([]types.Provider, 0, len(fallbackNames)+1)
	seen := make(map[string]bool)

	primaryName := primaryProvider.GetProviderName()
	providerChain = append(providerChain, primaryProvider)
	seen[primaryName] = true

	for _, fallbackName := range fallbackNames {
		if seen[fallbackName] {
			continue
		}

		fallbackProvider, err := manager.GetProvider(fallbackName)
		if err != nil {
			continue
		}

		providerChain = append(providerChain, fallbackProvider)
		seen[fallbackName] = true
	}

	return providerChain
}

func HandleStreamingCompletion(resolver app.ConfigResolver, c *gin.Context, provider types.Provider, request types.Completion) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	circuitBreakers := resolver.GetCircuitBreaker()
	providerName := provider.GetProviderName()
	circuitBreaker := circuitBreakers[providerName]

	chunks, err := provider.CompleteStream(c.Request.Context(), request.Messages)

	if err != nil {
		resolver.GetLogger().Error("Provider streaming failed", zap.Error(err))
		c.SSEvent("error", gin.H{
			"error": "Failed to start streaming completion",
		})
		return
	}

	// Stream chunks to client
	for chunk := range chunks {

		circuitBreaker.Execute(chunk.Error)

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
	// Extract context.Context once at the HTTP boundary
	ctx := c.Request.Context()

	var request types.Completion
	retry := resolver.GetRetry()
	circuitBreakers := resolver.GetCircuitBreaker()

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

	resolver.GetLogger().Info("Completion request received",
		zap.Int("message_count", len(request.Messages)),
		zap.String("model", request.Model),
		zap.Bool("stream", request.Stream),
	)

	router := resolver.GetRouter()

	providerStruct, err := router.SelectProvider(ctx, &types.SelectProviderInput{
		Messages: request.Messages,
		Circuits: circuitBreakers,
	})

	provider := providerStruct.Provider
	// model := providerStruct.Model

	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "no available providers, cannot process requests",
		})
		return
	}

	providerChain := buildProviderChain(provider, resolver.GetFallbackChain(), resolver.GetProviderManager())

	resolver.GetLogger().Info("Provider chain built",
		zap.Int("chain_length", len(providerChain)),
		zap.String("primary_provider", provider.GetProviderName()),
	)

	if request.Stream {
		HandleStreamingCompletion(resolver, c, provider, request)
		return
	}

	var lastErr error

	for i, currentProvider := range providerChain {
		currentProviderName := currentProvider.GetProviderName()
		currentCircuitBreaker := circuitBreakers[currentProviderName]

		resolver.GetLogger().Debug("Trying provider",
			zap.Int("attempt", i+1),
			zap.Int("total", len(providerChain)),
			zap.String("provider", currentProviderName),
			zap.String("circuit_state", currentCircuitBreaker.GetState()),
		)

		response, err := resilience.Do(ctx, currentProviderName, retry, func(ctx context.Context) (*types.Message, error) {
			return currentProvider.Complete(ctx, request.Messages)
		})

		currentCircuitBreaker.Execute(err)

		if err != nil {
			resolver.GetLogger().Warn("Provider failed, trying next in chain",
				zap.String("provider", currentProviderName),
				zap.Error(err),
				zap.Int("remaining_providers", len(providerChain)-i-1),
			)
			lastErr = err
			continue
		}

		resolver.GetLogger().Info("Provider succeeded",
			zap.String("provider", currentProviderName),
			zap.Int("attempt_number", i+1),
		)

		c.JSON(http.StatusOK, gin.H{
			"message":  response.Content,
			"role":     response.Role,
			"provider": currentProviderName,
		})
		return
	}

	resolver.GetLogger().Error("All providers in fallback chain failed",
		zap.Int("providers_tried", len(providerChain)),
		zap.Error(lastErr),
	)

	c.JSON(http.StatusInternalServerError, gin.H{
		"error":       "All providers in fallback chain failed",
		"last_error":  lastErr.Error(),
		"tried_count": len(providerChain),
	})

}

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
