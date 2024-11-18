---
title: <CMD NAME HERE>
description: description here in 2 to 3 lines
---

<!-- description in 2 to 3 sentences, following is an example -->
<!-- for sample template please check sample_command_docs.md file -->

## Syntax

```bash
# command syntax here
```

<!-- If the command have subcommands please mention but do not consider them as arguments -->
<!-- please mention them in subcommands section and create their individual documents -->

## Parameters

<!-- please add all parameters, small description, type and required, see example for SET command-->

| Parameter | Description | Type | Required |
| --------- | ----------- | ---- | -------- |
|           |             |      |          |
|           |             |      |          |
|           |             |      |          |

## Return values

<!-- add all scenarios, see below example for SET -->

| Condition | Return Value |
| --------- | ------------ |
|           |              |
|           |              |
|           |              |

## Behaviour

<!-- How does the command execute goes here, kind of explaining the underlying algorithm -->
<!-- see below example for SET command -->
<!-- Please modify for the command by going through the code -->

- Bullet 1
- Bullet 2
- Bullet 3

## Errors

<!-- sample errors, please update for commands-->
<!-- please add all the errors here -->
<!-- incase of a dynamic error message, feel free to use variable names -->
<!-- follow below bullet structure -->

1. `Wrong type of value or key`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when attempting to use the command on a key that contains a non-string value.

## Example Usage

<!-- please add examples for the command -->
<!-- Its good to have example using flags and arguments -->
<!-- Also adding for errors provide idea about the command -->

<!-- example heading in h3 -->

### Basic Usage

<!-- Always use bash code style block -->

```bash
127.0.0.1:7379> SET foo bar
OK
```

<!-- Please use detailed scenarios and edges cases if possible -->
<!-- example heading in h3 -->

### Invalid usage

```bash
127.0.0.1:7379> <CMD syntax here>
(error) ERR syntax error
```

<!-- Optional: Used when additional information is to conveyed to users -->
<!-- For example warnings about usage ex: Keys * -->
<!-- OR alternatives of the commands -->
<!-- Or perhaps deprecation warning -->
<!-- anything related to the command which cannot be shared in other sections -->

<!-- Optional -->

## Best Practices

<!-- Optional -->

## Alternatives

<!-- Optional -->

## Notes

<!-- Optional -->

## Subcommands

<!-- if the command you are working on has subcommands -->
<!-- please mention them here and add links to the pages -->
<!-- please see below example for COMMAND docs -->
<!-- follow below bullet structure -->

- **subcommand**: Optional. Available subcommands include:
  - `subcommand1` : some description of the subcommand.
