FROM golang:1.24 AS builder
WORKDIR /src
COPY . .

# Detect platform and set architecture dynamically
# Had to do this as I have a arm64 machine
ARG TARGETPLATFORM
RUN case "$TARGETPLATFORM" in \
      "linux/amd64") export GOARCH=amd64 ;; \
      "linux/arm64") export GOARCH=arm64 ;; \
      *) echo "Unsupported platform: $TARGETPLATFORM" && exit 1 ;; \
    esac && \
    CGO_ENABLED=0 GOOS=linux go build -o /src/dicedb

FROM alpine:latest
COPY --from=builder /src/dicedb /src/dicedb
COPY VERSION .
RUN chmod +x /src/dicedb

EXPOSE 7379

ENTRYPOINT ["/src/dicedb"]
CMD []
