---
name: Bug in DiceDB command vs Redis TCL Test
about: Use this template to report any issue in command behaviour vs Redis
title: 'Inconsistent `{CMD}`: <Describe the error in one concise line>'
labels: ''
assignees: ''

---

## Steps to reproduce

{steps_to_reproduce}

## Expected output

The expected output when the above set of commands when run on Redis

```
{expected_output}
```

## Observed output

The observed output when the above set of commands when run on DiceDB

```
{observed_output}
```

The steps to run the test cases are mentioned in the README of the [dice-tests repository](https://github.com/AshwinKul28/dice-tests).


## Expectations for resolution

This issue will be considered resolved when the following things are done

1. changes in the [`dice`](https://github.com/dicedb/dice) code to meet the expected behavior
2. Successful run of the tcl test behavior

You can find the tests under the `tests` directory of the [`dice`](https://github.com/dicedb/dice) repository and the steps to run are in the [README file](https://github.com/dicedb/dice). Refer to the following links to set up DiceDB and Redis 7.2.5 locally

- [setup DiceDB locally](https://github.com/dicedb/dice)
- [setup Redis 7.2.5 locally](https://gist.github.com/arpitbbhayani/94aedf279349303ed7394197976b6843)
- [setup DiceDB CLI](https://github.com/dicedb/dice)
