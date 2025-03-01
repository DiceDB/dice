---
name: Issue with Base Instructions
about: This template is supposed to be used by maintainers to create issue
title: ''
labels: ''
assignees: ''
---

> Issue details go here ...

## Setup Instructions

1. setup DiceDB server locally from the source - [instructions](https://github.com/dicedb/dice)
2. setup DiceDB Go SDK locally from the source - [instructions](https://github.com/dicedb/dicedb-go)
3. setup DiceDB CLI locally from the source - [instructions](https://github.com/dicedb/dicedb-cli)
4. refer to the `Pointing to local checked-out `dicedb-go` section in `README`.

### Start the DiceDB server

```
$ go run main.go --log-level debug
```

## Follow the contribution guidelines

These are general guidelines to follow before you submit a patch. Please mark them as done once you complete them

- [ ] please go through the [CONTRIBUTING](https://github.com/DiceDB/dice/tree/master/CONTRIBUTING) guide
- [ ] follow [LOGGING best practices](https://github.com/DiceDB/dice/blob/master/CONTRIBUTING/logging.md)
- [ ] follow [Golang best practices](https://github.com/DiceDB/dice/blob/master/CONTRIBUTING/go.md)
- [ ] run `make lint` on your local copy of the codebase
