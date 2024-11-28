# Build Stage
FROM golang:1.23 AS builder

WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy application source code and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# Runtime Stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates

# Copy the built application
WORKDIR /root/
COPY --from=builder /app/server .

# Expose the application port
EXPOSE 8080

# Start the application
CMD ["./server"]
