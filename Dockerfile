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
# - AGENT_NAME, AGENT_VERSION
# - TOOL_NAME, TOOL_DESCRIPTION
# - LLM_PROVIDER, LLM_MODEL, LLM_API_KEY
# - See examples directory for complete configuration examples

# Expose port for HTTP mode
EXPOSE 3000

# Default command: run in daemon mode
CMD ["/app/speelka-agent"]