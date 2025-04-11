FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies needed for building
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o speelka-agent ./cmd/server/main.go

# Final stage: create a minimal image
FROM alpine:latest

WORKDIR /app

# Install CA certificates for HTTPS connections
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/speelka-agent /app/speelka-agent

# Set executable permissions
RUN chmod +x /app/speelka-agent

# Configuration is provided through environment variables
# Examples:
# - SPL_AGENT_NAME, SPL_AGENT_VERSION
# - SPL_TOOL_NAME, SPL_TOOL_DESCRIPTION
# - SPL_LLM_PROVIDER, SPL_LLM_MODEL, SPL_LLM_API_KEY
# - SPL_RUNTIME_LOG_LEVEL, SPL_RUNTIME_LOG_OUTPUT, SPL_RUNTIME_STDIO_ENABLED, SPL_RUNTIME_STDIO_BUFFER_SIZE
# - See README.md for complete configuration options

# Expose port for HTTP mode
EXPOSE 3000

# Default command: run in daemon mode
CMD ["/app/speelka-agent"]