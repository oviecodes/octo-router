# Octo Router

A production-ready, open-source LLM router built in Go. Route requests across multiple LLM providers with intelligent load balancing, provider-specific configurations, and comprehensive input validation.

## Features

- **Multi-Provider Support**: OpenAI, Anthropic (Claude), and Google Gemini
- **Standardized Model Naming**: Consistent `provider/model` format with built-in pricing metadata
- **Intelligent Routing**:
  - **Cost-Based Routing**: Automatically select the cheapest model based on tier constraints
  - **Latency-Based Routing**: Auto-select fastest provider using real-time EMA tracking
  - **Weighted Routing**: Distribute traffic based on percentage weights (e.g., 70/20/10)
  - **Round-Robin**: Distribute load evenly across providers
  - **Tier-Based Selection**: Control quality/cost trade-offs with tier constraints (budget, standard, premium, ultra-premium)
- **Fallback Chain**: Automatic failover to backup providers when primary provider fails
- **Provider-Specific Defaults**: Configure model and token limits per provider
- **Input Validation**: Comprehensive request validation with detailed error messages
- **Token Estimation**: Local token counting using tiktoken (no API calls)
- **Cost Tracking**: Automatic per-request cost calculation with Prometheus metrics
- **Resilience**: Circuit breakers, retries with exponential backoff, and error translation
- **Dynamic Provider Management**: Enable/disable providers without code changes
- **Health Monitoring**: Built-in health check endpoint
- **Structured Logging**: Production-ready logging with zap
- **Streaming Support**: Server-Sent Events for streaming responses

## Supported Providers

All models use the standardized `provider/model` naming format.

| Provider      | Status    | Models                                                                                                     |
| ------------- | --------- | ---------------------------------------------------------------------------------------------------------- |
| OpenAI        | Supported | openai/gpt-5, openai/gpt-5.1, openai/gpt-4o, openai/gpt-4o-mini, openai/gpt-3.5-turbo                      |
| Anthropic     | Supported | anthropic/claude-opus-4.5, anthropic/claude-sonnet-4, anthropic/claude-haiku-4.5, anthropic/claude-haiku-3 |
| Google Gemini | Supported | gemini/gemini-2.5-flash, gemini/gemini-2.5-flash-lite, gemini/gemini-2.0-pro                               |

See [Model Standardization](docs/MODEL_STANDARDIZATION.md) for complete pricing and model details.

## Getting Started

### Prerequisites

- Go 1.21 or higher
- API keys for desired providers

### Installation

```bash
git clone https://github.com/oviecodes/octo-router.git
cd octo-router
go mod download
```

### Configuration

Create a `config.yaml` file in the project root:

```yaml
providers:
  - name: openai
    apiKey: ${OPENAI_API_KEY}
    enabled: true

  - name: anthropic
    apiKey: ${ANTHROPIC_API_KEY}
    enabled: true

  - name: gemini
    apiKey: ${GEMINI_API_KEY}
    enabled: true

models:
  defaults:
    openai:
      model: "openai/gpt-4o-mini"
      maxTokens: 4096

    anthropic:
      model: "anthropic/claude-sonnet-4"
      maxTokens: 4096

    gemini:
      model: "gemini/gemini-2.5-flash"
      maxTokens: 8192

routing:
  strategy: weighted # cost-based, round-robin, latency-based, weighted
  
  # Weights (must sum to any number, normalized automatically)
  weights:
    openai: 70
    anthropic: 20
    gemini: 10

  fallbacks:
    - anthropic
    - gemini
    - openai

resilience:
  timeout: 30000

  retries:
    maxAttempts: 3
    initialDelay: 1000
    maxDelay: 10000
    backoffMultiplier: 2

  circuitBreaker:
    failureThreshold: 5
    resetTimeout: 60000

cache:
  enabled: true
  ttl: 3600
```

**Note**: Use environment variables for API keys. The router supports `${VAR_NAME}` syntax for environment variable substitution.

### Running the Server

```bash
go run main.go
```

The server starts on `localhost:8000` by default.

## API Reference

### Health Check

Check the router's health and number of enabled providers.

```bash
GET /health
```

**Response:**

```json
{
  "status": "healthy",
  "providers": 3
}
```

### Chat Completions

Send messages to LLM providers via the router.

```bash
POST /v1/chat/completions
```

**Request Body:**

```json
{
  "messages": [
    {
      "role": "user",
      "content": "What is the capital of France?"
    }
  ],
  "model": "openai/gpt-4o-mini",
  "temperature": 0.7,
  "max_tokens": 1000,
  "stream": false
}
```

**Parameters:**

| Field               | Type    | Required | Description                                        |
| ------------------- | ------- | -------- | -------------------------------------------------- |
| `messages`          | array   | Yes      | Array of message objects with `role` and `content` |
| `model`             | string  | No       | Override default model for selected provider       |
| `temperature`       | float   | No       | Sampling temperature (0-2)                         |
| `max_tokens`        | integer | No       | Maximum tokens to generate (1-100000)              |
| `top_p`             | float   | No       | Nucleus sampling (0-1)                             |
| `frequency_penalty` | float   | No       | Frequency penalty (-2 to 2)                        |
| `presence_penalty`  | float   | No       | Presence penalty (-2 to 2)                         |
| `stream`            | boolean | No       | Enable streaming responses (not yet implemented)   |

**Message Roles:**

- `user`: User messages
- `assistant`: Assistant responses (for conversation history)
- `system`: System instructions

**Response:**

```json
{
  "message": "The capital of France is Paris.",
  "role": "assistant",
  "provider": "*providers.OpenAIProvider"
}
```

**Error Response:**

```json
{
  "error": "Validation failed",
  "details": [
    {
      "field": "messages[0].role",
      "message": "Role must be one of: user assistant system"
    }
  ]
}
```

### Admin Endpoints

#### Get All Providers Configuration

```bash
POST /admin/config
```

**Response:**

```json
{
  "providers": [
    {
      "name": "openai",
      "apiKey": "sk-***",
      "enabled": true
    }
  ]
}
```

#### Get Enabled Providers

```bash
POST /admin/providers
```

**Response:**

```json
{
  "enabled": [
    {
      "name": "openai",
      "apiKey": "sk-***",
      "enabled": true
    }
  ],
  "count": 3
}
```

## Architecture

### Project Structure

```
llm-router/
├── cmd/
│   └── internal/
│       ├── providers/      # Provider implementations
│       │   ├── provider.go
│       │   ├── openai.go
│       │   ├── anthropic.go
│       │   └── gemini.go
│       ├── router/         # Routing logic
│       │   └── router.go
│       └── server/         # HTTP handlers
│           ├── server.go
│           └── validation.go
├── config/                 # Configuration loading
│   └── config.go
├── types/                  # Shared types
│   ├── completion.go
│   ├── message.go
│   ├── provider.go
│   └── router.go
├── utils/                  # Utility functions
├── config.yaml            # Configuration file
└── main.go
```

### How It Works

1. **Configuration Loading**: At startup, the router loads provider configurations, model defaults, and fallback chain from `config.yaml`
2. **Provider Initialization**: Enabled providers are initialized with their respective API clients
3. **Model Validation**: Model names are validated against the centralized model catalog with pricing metadata
4. **Request Handling**: Incoming requests are validated, routed to a provider using the configured strategy, and responses are normalized
5. **Fallback Chain**: If a provider fails, the router automatically tries fallback providers in the configured order
6. **Error Handling**: Provider-specific errors are translated to domain errors with retry logic
7. **Resilience**: Circuit breakers track provider health, retries handle transient failures
8. **Metrics**: Prometheus metrics track requests, latency, costs, and circuit breaker state

### Routing Strategies

Currently supported:

- **Round-robin**: Distributes requests evenly across enabled providers
- **Cost-based routing**: Pick the cheapest model for a request depending on the tier set in config.yaml
- **Latency-based routing**: Dynamically routes to the fastest provider using Exponential Moving Average (EMA) tracking
- **Weighted routing**: Routes traffic based on user-defined weights (e.g., A=80%, B=20%)

Planned:

- Provider-specific routing rules

### Fallback Chain

The fallback chain provides automatic failover when the primary provider fails, ensuring high availability and reliability.

**How It Works:**

1. **Provider Selection**: The router selects a primary provider using the configured routing strategy (e.g., round-robin)
2. **Fallback Chain Building**: A fallback chain is built by combining the primary provider with the configured fallback providers from `config.yaml`
3. **Sequential Retry**: If the primary provider fails, the router automatically tries each fallback provider in order
4. **Circuit Breaker Integration**: Each provider's circuit breaker state is checked and updated during the fallback process
5. **Deduplication**: The chain automatically prevents duplicate providers (e.g., if primary is already in fallback list)
6. **Immediate Success Return**: The first successful provider response is returned immediately
7. **All-Failed Error**: Only returns an error if all providers in the chain fail

**Configuration:**

```yaml
routing:
  strategy: round-robin
  fallbacks:
    - anthropic # Try Anthropic first if primary fails
    - gemini # Then try Gemini
    - openai # Finally try OpenAI
```

**Example Flow:**

1. Round-robin selects `openai` as primary
2. Fallback chain built: `[openai, anthropic, gemini]`
3. OpenAI fails → automatically retries with Anthropic
4. Anthropic succeeds → returns response immediately
5. Circuit breaker updated for both OpenAI (failure) and Anthropic (success)

**Benefits:**

- **High Availability**: Requests succeed even when providers are down
- **Transparent Failover**: Clients don't need to handle provider failures
- **Cost Optimization**: Configure cheaper providers as fallbacks
- **Latency Management**: Try faster providers before slower ones
- **Comprehensive Logging**: Detailed logs track each fallback attempt with provider names, error details, and remaining providers

**Note**: Fallback chain works for both streaming and non-streaming completions. Each provider in the chain respects the configured retry policy and circuit breaker settings.

## Validation

The router performs comprehensive input validation:

- **Message validation**: Required fields, role validation, content length limits
- **Parameter validation**: Range checks for temperature, max_tokens, penalties
- **Business logic validation**: First message role requirements, total content size limits
- **Detailed error messages**: Clear, actionable error messages for clients

## Token Counting

Token counting uses tiktoken for local estimation:

- No API calls required
- No rate limiting
- Fast and accurate for cost estimation
- Works across all providers (normalized to OpenAI's encoding)

## Development

### Running Tests

```bash
go test ./...
```

### Environment Variables

```bash
export OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="sk-..."
export GEMINI_API_KEY="..."
export APP_ENV="development"  # or "production"
```

### Adding a New Provider

1. Create provider implementation in `cmd/internal/providers/`
2. Implement the `Provider` interface:
   ```go
   type Provider interface {
       Complete(ctx context.Context, messages []types.Message) (*types.Message, error)
       CountTokens(ctx context.Context, messages []types.Message) (int, error)
   }
   ```
3. Add provider case to `ConfigureProviders()` in `provider.go`
4. Add default configuration to `config.yaml`

## Roadmap

- [x] Streaming support (Server-Sent Events)
- [x] Proper error handling for different types of Error
- [x] Circuit breaker for provider failures
- [x] Fallback chain for automatic provider failover
- [x] Model standardization with pricing metadata
- [x] Metrics and observability (Prometheus)
- [ ] Request/response caching (semantic caching planned)
- [ ] Cost tracking and reporting (cost calculation implemented)
- [ ] Rate limiting per provider
- [x] Custom routing strategies
- [ ] Function/Tool calling support
- [ ] Multi-tenancy support

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

<!-- MIT License - see LICENSE file for details -->

## Acknowledgments

Built with:

- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [Zap Logger](https://github.com/uber-go/zap)
- [Viper Configuration](https://github.com/spf13/viper)
- [tiktoken-go](https://github.com/pkoukk/tiktoken-go)
