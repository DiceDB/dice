.PHONY: build test build-docker

build:
	CGO_ENABLED=0 GOOS=linux go build -o ./dice-server

test:
	-./runtest --verbose --tags -slow --dump-logs

build-docker:
	docker build --tag dicedb/dice-server:latest --tag dicedb/dice-server:0.0.1 .
