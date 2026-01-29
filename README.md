# Octo Router

**A production-ready, open-source AI Gateway built in Go.**

Octo Router optimizes your LLM infrastructure by intelligently routing requests across providers (OpenAI, Anthropic, Gemini) to reduce costs, improve latency, and ensure high availability.

[**Read the Full Documentation**](docs/content/docs/index.mdx)

## Key Features

- **Smart Routing**: Route based on cost, latency, or custom weights.
- **Semantic Routing**: Locally classify user intent (coding vs. chat) using ONNX to pick the best model.
- **Cost Management**: Set daily budgets, track spending, and auto-downgrade to cheaper models.
- **Resilience**: Built-in circuit breakers, retries, and automatic fallbacks.
- **High Performance**: Written in Go with Redis-backed state management.
- **Cloud Native**: Docker-ready and easy to deploy on Kubernetes.

## Quick Start

### Docker (Recommended)

```bash
# Start the router and Redis
docker compose up --build
```

The server will start at `http://localhost:8000`.

### Local Development

1. **Install Dependencies**: Go 1.25+, Redis, and ONNX Runtime.
2. **Clone & Run**:
   ```bash
   git clone https://github.com/oviecodes/octo-router.git
   go mod download
   go run cmd/server/main.go
   ```

## Documentation

- [Getting Started](docs/content/docs/getting-started.mdx)
- [Configuration Reference](docs/content/docs/configuration.mdx)
- [API Reference](docs/content/docs/api-reference.mdx)
- [Semantic Routing](docs/content/docs/semantic-routing.mdx)

## License

MIT
