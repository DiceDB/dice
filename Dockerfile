FROM golang:1.24 AS builder
WORKDIR /dicedb
COPY go.mod go.sum ./
RUN go mod download -x
COPY . .
RUN make build

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /dicedb/dicedb ./
EXPOSE  7379

ENTRYPOINT ["/app/dicedb"]
