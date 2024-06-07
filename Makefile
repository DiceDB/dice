.PHONY: build-docker

build-docker:
	docker build --tag dicedb/dice-server:latest --tag dicedb/dice-server:0.0.1 .
