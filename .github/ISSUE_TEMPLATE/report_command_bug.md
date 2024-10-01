---
name: Bug in DiceDB command
about: Use this template to report bug in any command behaviour vs Redis
title: 'Inconsistent `{CMD}`: <Describe the error in one concise line>'
labels: ''
assignees: ''

---

## Steps to reproduce

{steps_to_reproduce}

## Expected output

The expected output when the above set of commands (maybe when run on Redis)

```
{expected_output}
```

## Observed output

The observed output when the above set of commands when run on DiceDB

```
{observed_output}
```

## Expectations for resolution

This issue will be considered resolved when the following things are done

1. changes in the [`dice`](https://github.com/dicedb/dice) code to meet the expected behavior
2. addition of relevant test case to ensure we catch the regression

You can find the tests under the `integration_tests` directory of the [`dice`](https://github.com/dicedb/dice) repository and the steps to run are in the [README file](https://github.com/dicedb/dice). Refer to the following links to set up DiceDB and Redis 7.2.5 locally

- [setup DiceDB locally](https://github.com/dicedb/dice)
- [setup Redis 7.2.5 locally](https://gist.github.com/arpitbbhayani/94aedf279349303ed7394197976b6843)
- [setup DiceDB CLI](https://github.com/dicedb/dice)
