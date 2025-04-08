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
COPY --from=builder /app/examples /app/examples

# Set executable permissions
RUN chmod +x /app/speelka-agent

# Set environment variables
ENV CONFIG_JSON=""

# Expose port for HTTP mode
EXPOSE 3000

# Set the entrypoint
ENTRYPOINT ["/app/speelka-agent"]

# Default command: run in daemon mode
CMD ["--daemon"]