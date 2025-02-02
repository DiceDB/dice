---
name: Report a bug [maintainers only]
about: This template is supposed to be used by maintainers to report a bug
title: 'Bug: '
labels: ''
assignees: ''
---

> Pick this issue only after #{PR_NUM} is merged.

{details of the issue}

## Steps to patch

1. setup DiceDB server locally from source - [instructions](https://github.com/dicedb/dice)
2. setup DiceDB CLI locally from source - [instructions](https://github.com/dicedb/dice)

## Follow the contribution guidelines

These are general guidelines to follow before you submit a patch. Please mark them as done
once you complete them

- [ ] please go through the [CONTRIBUTING](https://github.com/DiceDB/dice/tree/master/CONTRIBUTING) guide
- [ ] follow [LOGGING best practices](https://github.com/DiceDB/dice/blob/master/CONTRIBUTING/logging.md)
- [ ] follow [Golang best practices](https://github.com/DiceDB/dice/blob/master/CONTRIBUTING/go.md)
- [ ] run `make lint` on your local copy of the codebase
