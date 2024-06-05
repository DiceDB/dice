# syntax=docker/dockerfile:1

FROM golang:1.17

WORKDIR /dicedb
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /dicedb/dice
EXPOSE 7379
CMD ["/dicedb/dice"]
