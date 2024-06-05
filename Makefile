.PHONY: build-docker

build-docker:
	docker build --tag dicedb/dice:latest --tag dicedb/dice:0.0.1 .
