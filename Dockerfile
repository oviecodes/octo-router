# Build stage
FROM golang:bookworm AS builder

# Install build dependencies
RUN apt-get update && apt-get install -y gcc libc6-dev wget tar

WORKDIR /app

# Download ONNX Runtime library
# We download it in the builder stage to keep the final image slim
RUN ARCH=$(uname -m) && \
    if [ "$ARCH" = "x86_64" ]; then \
    ONNX_ARCH="x64"; \
    elif [ "$ARCH" = "aarch64" ]; then \
    ONNX_ARCH="aarch64"; \
    else \
    echo "Unsupported architecture: $ARCH"; exit 1; \
    fi && \
    wget https://github.com/microsoft/onnxruntime/releases/download/v1.23.2/onnxruntime-linux-$ONNX_ARCH-1.23.2.tgz \
    && tar -xzf onnxruntime-linux-$ONNX_ARCH-1.23.2.tgz \
    && mv onnxruntime-linux-$ONNX_ARCH-1.23.2/lib/libonnxruntime.so.1.23.2 /usr/local/lib/libonnxruntime.so \
    && rm -rf onnxruntime-linux-$ONNX_ARCH-1.23.2*

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

RUN go mod tidy

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o main ./cmd/server/main.go

# Run stage
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y ca-certificates libc6 libgomp1

WORKDIR /app

RUN mkdir -p /go/pkg/mod/github.com/sugarme/tokenizer@v0.3.0/

# Copy binary from builder
COPY --from=builder /app/main .
COPY --from=builder /usr/local/lib/libonnxruntime.so /usr/local/lib/libonnxruntime.so
COPY --from=builder /go/pkg/mod/github.com/sugarme/tokenizer@v0.3.0/pretrained /go/pkg/mod/github.com/sugarme/tokenizer@v0.3.0/pretrained

# Copy config and assets
COPY config.yaml .
COPY assets/ ./assets/

# Command to run
CMD ["./main"]
