FROM golang:1.21 AS builder
WORKDIR /dicedb
COPY go.mod go.sum ./
RUN go mod download -x
COPY . .
RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags "-s -w" -o /dicedb/dice-server

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /dicedb/dice-server ./
EXPOSE  7379
CMD ["/app/dice-server"]
