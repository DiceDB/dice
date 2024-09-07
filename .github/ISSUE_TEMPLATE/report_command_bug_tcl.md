---
name: Bug in DiceDB command vs Redis TCL Test
about: Use this template to report any issue in command behaviour vs Redis
title: 'Inconsistent `{CMD}`: <Describe the error in one concise line>'
labels: ''
assignees: ''

---

There is an inconsistent and incorrect behavior (probably a bug) in the current implementation of the `{CMD}` command and here are the details you need you reproduce it and fix it. 

## Steps to reproduce

```
{steps_to_reproduce}
```

## Expected output

The expected output when the above set of commands when run on Redis

```
{expected_output}
```

You can find the related TCL test case - {tcl_testcase_link}

This test case should pass once this issue is fixed. The steps to run the test cases are mentioned
in the README of the [dice-tests repository](https://github.com/AshwinKul28/dice-tests/blob/main/README.md).

## Observed output

The observed output when the above set of commands when run on DiceDB

```
{observed_output}
```

## Expectations for resolution

This issue will be considered resolved when the following things are done

1. changes in the [`dice`](https://github.com/dicedb/dice) code to meet the expected behavior
2. addition of new unit or integration tests in the [`dice`](https://github.com/dicedb/dice) repository

You can find the tests under the `tests` directory of the [`dice`](https://github.com/dicedb/dice) repository and the steps to run are in the [README file](https://github.com/dicedb/dice).
