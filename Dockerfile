FROM golang:1.25-alpine AS builder

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
# - SPL_AGENT_TOOL_NAME, SPL_AGENT_TOOL_DESCRIPTION
# - SPL_AGENT_LLM_PROVIDER, SPL_AGENT_LLM_MODEL, SPL_AGENT_LLM_APIKEY
# - SPL_RUNTIME_LOG_DEFAULTLEVEL, SPL_RUNTIME_LOG_FORMAT, SPL_RUNTIME_TRANSPORTS_STDIO_ENABLED, SPL_RUNTIME_TRANSPORTS_STDIO_BUFFERSIZE
# - See README.md for complete configuration options

# Expose port for HTTP mode
EXPOSE 3000

# Default command: run in daemon mode
CMD ["/app/speelka-agent"]