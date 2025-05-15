# Start from a slim alpine image
FROM golang:1.24.2-alpine AS builder
WORKDIR /src
COPY . .

# Install musl-dev to make sure we get the required libraries
RUN apk add --no-cache gcc musl-dev libc6-compat make
RUN make build

FROM alpine
WORKDIR /src

# Copy the Go binary from the builder stage
COPY --from=builder /usr/local/go /usr/local/go
ENV PATH="/usr/local/go/bin:${PATH}"

COPY --from=builder /src/dicedb /src/VERSION /src/

EXPOSE 7379

ENTRYPOINT ["/src/dicedb"]
CMD []