FROM golang:1.23 AS builder
WORKDIR /dicedb
COPY go.mod go.sum ./
RUN go mod download -x
COPY . .
RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags "-s -w" -o /dicedb/dicedb

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /dicedb/dicedb ./
EXPOSE  7379

ENTRYPOINT ["/app/dicedb"]
