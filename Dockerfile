# Start from a slim apline image
FROM golang:1.24.1-alpine AS builder
WORKDIR /src
COPY . .

# Install musl-dev to make sure we get the required libraries
RUN apk add --no-cache gcc musl-dev libc6-compat make
RUN make build

FROM alpine
COPY --from=builder /src/dicedb /src/VERSION /src/

EXPOSE 7379

ENTRYPOINT ["/src/dicedb"]
CMD []
