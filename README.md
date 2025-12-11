# Octo Router

A production-ready, open-source LLM router built in Go. Route requests across multiple LLM providers with intelligent load balancing, provider-specific configurations, and comprehensive input validation.

## Features

- **Multi-Provider Support**: OpenAI, Anthropic (Claude), and Google Gemini
- **Flexible Routing**: Round-robin load balancing with support for custom routing strategies
- **Provider-Specific Defaults**: Configure model and token limits per provider
- **Input Validation**: Comprehensive request validation with detailed error messages
- **Token Estimation**: Local token counting using tiktoken (no API calls)
- **Dynamic Provider Management**: Enable/disable providers without code changes
- **Health Monitoring**: Built-in health check endpoint
- **Structured Logging**: Production-ready logging with zap

## Supported Providers

| Provider      | Status    | Models                                           |
| ------------- | --------- | ------------------------------------------------ |
| OpenAI        | Supported | gpt-4, gpt-4o, gpt-4o-mini, gpt-3.5-turbo        |
| Anthropic     | Supported | claude-3-5-sonnet, claude-3-haiku, claude-3-opus |
| Google Gemini | Supported | gemini-2.5-flash, gemini-1.5-flash, gemini-3-pro |

## Getting Started

### Prerequisites

- Go 1.21 or higher
- API keys for desired providers

### Installation

```bash
git clone https://github.com/yourusername/llm-router.git
cd llm-router
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
    - name: openai
      model: "gpt-4o-mini"
      maxTokens: 4096

    - name: anthropic
      model: "claude-3-5-sonnet-20241022"
      maxTokens: 4096

    - name: gemini
      model: "gemini-2.5-flash"
      maxTokens: 8192

routing:
  strategy: round-robin
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
  "model": "gpt-4o-mini",
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
├── docs/                   # Documentation
│   ├── multi-tenancy-implementation.md
│   └── streaming-implementation.md
├── config.yaml            # Configuration file
└── main.go
```

### How It Works

1. **Configuration Loading**: At startup, the router loads provider configurations and model defaults from `config.yaml`
2. **Provider Initialization**: Enabled providers are initialized with their respective API clients
3. **Request Handling**: Incoming requests are validated, routed to a provider using the configured strategy, and responses are normalized
4. **Error Handling**: Validation errors and provider failures are handled gracefully with structured error responses

### Routing Strategies

Currently supported:

- **Round-robin**: Distributes requests evenly across enabled providers

Planned:

- Cost-based routing
- Latency-based routing
- Provider-specific routing rules

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

- Streaming support (Server-Sent Events) [Done]
- [ ] Cost tracking and reporting
- [ ] Rate limiting per provider
- [ ] Circuit breaker for provider failures
- [ ] Multi-tenancy support
- [ ] Custom routing strategies
- [ ] Metrics and observability (Prometheus)
- [ ] Request/response caching
- [ ] Function/Tool calling support

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
