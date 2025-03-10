# Stage 1: Build the binary
FROM golang:1.24 AS builder
WORKDIR /src

# Copy source files
COPY . .

# Build the binary
RUN make build

# Stage 2: Create a minimal final image
FROM debian:bullseye-slim
WORKDIR /
COPY --from=builder /src/dicedb /src/dicedb

# Ensure the binary has execute permissions (if needed)
# RUN chmod +x /dicedb

# Expose the required port
EXPOSE 7379

COPY VERSION .

# Set the entrypoint to run the binary
ENTRYPOINT ["/src/dicedb"]