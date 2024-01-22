FROM golang:1.18 AS builder

RUN apt-get update && apt-get -y dist-upgrade
RUN apt install -y netcat

ENV GO111MODULE=on

RUN mkdir -p /go/src/github.com/dicedb/dice

WORKDIR /go/src/github.com/dicedb/dice

ADD . /go/src/github.com/dicedb/dice
# Build
COPY ./go.mod .
COPY ./go.sum .

RUN go mod download

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o dice main.go 


FROM alpine:latest

WORKDIR /app/server

COPY --from=builder /go/src/github.com/dicedb/dice ./

EXPOSE 7379

RUN chmod +x ./dice

CMD [ "./dice" ]


