---
name: Inconsistency DiceDB w.r.t Redis
about: Use this template to report any inconsistency you observed in DiceDB w.r.t Redis
title: 'Inconsistent `{CMD}`: <Describe the error in one concise line>'
labels: ''
assignees: ''

---

The command `{CMD}` is not consistent with Redis implementation. Here are the steps to reproduce the issue

```
{steps_to_reproduce}
```

Here's the output I observed in Redis v7.2.5

```
{redis_output}
```

and here's the output I observed in DiceDB's latest commit of the `master` branch

```
{dicedb_output}
```

Make the implementation consistent with the Redis implementation.
Make sure you are using Redis version 7.2.5 as a reference for the
command implementation and to setup Redis

- [from source code](https://gist.github.com/arpitbbhayani/94aedf279349303ed7394197976b6843), or
- [use Docker](https://hub.docker.com/_/redis).
