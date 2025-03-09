FROM golang:1.24 AS builder
WORKDIR /src
COPY . .
RUN make build

RUN chmod +x /src/dicedb

EXPOSE 7379

ENTRYPOINT ["/src/dicedb"]
CMD []
