.PHONY: build test build-docker run test-one

build:
	CGO_ENABLED=0 GOOS=linux go build -o ./dice-server

format:
	go fmt ./...

run:
	go run main.go

test:
	go test -v -count=1 ./tests/

test-one:
	go test -v -count=1 --run $(TEST_FUNC) ./tests/...

unittest:
	go test -v -count=1 ./core/...

unittest-one:
	go test -v -count=1 --run $(TEST_FUNC) ./core/...

build-docker:
	docker build --tag dicedb/dice-server:latest --tag dicedb/dice-server:0.0.1 .
