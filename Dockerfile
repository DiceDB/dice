FROM golang:1.22.5 AS builder
WORKDIR /dicedb
COPY go.mod go.sum ./
RUN go mod download -x
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags "-s -w" -o /dicedb/dice-server

FROM --platform=linux/amd64 gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /dicedb/dice-server ./
CMD ["/app/dice-server"]
