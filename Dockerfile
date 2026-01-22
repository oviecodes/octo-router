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
    wget https://github.com/microsoft/onnxruntime/releases/download/v1.16.3/onnxruntime-linux-$ONNX_ARCH-1.16.3.tgz \
    && tar -xzf onnxruntime-linux-$ONNX_ARCH-1.16.3.tgz \
    && mv onnxruntime-linux-$ONNX_ARCH-1.16.3/lib/libonnxruntime.so.1.16.3 /usr/local/lib/libonnxruntime.so \
    && rm -rf onnxruntime-linux-$ONNX_ARCH-1.16.3*

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o main ./cmd/server/main.go

# Run stage
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y ca-certificates libc6 libgomp1

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .
COPY --from=builder /usr/local/lib/libonnxruntime.so /usr/local/lib/libonnxruntime.so

# Copy config and assets
COPY config.yaml .
COPY assets/ ./assets/

# Command to run
CMD ["./main"]
