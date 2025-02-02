---
name: Port command to IronHawk engine
about: This template is supposed to be used by maintainers tracking the port of a command to IronHawk
title: 'IronHawk Port: '
labels: ''
assignees: ''
---

We reimplemented the core of the DiceDB engine and re-wrote - the wire protocol, execution engine, and config management. This rewrite helped us gain 32% throughput over our existing benchmark. One of the core principles we focussed on was to make code easy to extend and debug. As an effort, we need to migrate the command from the old engine to IronHawk.

Here are the exact things to be taken care of to migrate command `{CMD}`

1. setup DiceDB server locally from the source - [instructions](https://github.com/dicedb/dice)
2. setup DiceDB CLI locally from the source - [instructions](https://github.com/dicedb/dice)

## Steps to execute

### Start the DiceDB server with IronHawk engine

```
$ go run main.go --engine ironhawk --log-level debug
```

### Start the DiceDB CLI with the IronHawk engine

```
$ go run main.go --engine ironhawk
```

## Porting the command

1. Find the current implementation of the command, the name of the function will be `eval{CMD}`. ex: `evalSET`, `evalGET`, etc. Most of them are present in the `store_eval.go` file.
2. Create a new file `internal/cmd/cmd_{cmd}.go` and follow the structure as `cmd_get.go`, `cmd_set.go`, and `cmd_ping.go` files.
3. Reimplement the old `eval{CMD}` function into a new file, Make a note of the return values of the new function.
4. If you think the implementation is complex, feel free to simplify
5. Document the code you have written and make sure it adheres to existing standard
6. Add `TODO` in the comment, if you feel there are things that need to be implemented later
7. Cover all possible cases for the command implementation
8. Do not delete the old implementation of the `eval` function

*No need* to write test cases for this new implementation. We will take care of this in one shot later. If the test fails, it is okay.

If you find any other bug while you are implementing it, you can either

1. fix it yourself and submit it in a new PR
2. raise a [GitHub issue](https://github.com/DiceDB/dice/issues)

## Follow the contribution guidelines

These are general guidelines to follow before you submit a patch. Please mark them as done
once you complete them

- [ ] please go through the [CONTRIBUTING](https://github.com/DiceDB/dice/tree/master/CONTRIBUTING) guide
- [ ] follow [LOGGING best practices](https://github.com/DiceDB/dice/blob/master/CONTRIBUTING/logging.md)
- [ ] follow [Golang best practices](https://github.com/DiceDB/dice/blob/master/CONTRIBUTING/go.md)
- [ ] run `make lint` on your local copy of the codebase
