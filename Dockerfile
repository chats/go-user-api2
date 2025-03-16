FROM golang:1.24-alpine AS builder

# Install build tools
RUN apk add --no-cache gcc musl-dev git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o go-user-api ./cmd

# Create a minimal runtime image
FROM alpine:latest

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/go-user-api .

# Copy .env file
COPY --from=builder /app/.env.example ./.env

# Expose ports
EXPOSE 8080 50051

# Create health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/api/health || exit 1

# Set entrypoint
ENTRYPOINT ["./go-user-api"]