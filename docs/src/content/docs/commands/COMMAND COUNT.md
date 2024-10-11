---
title: COMMAND COUNT
description: Documentation for the DiceDB command COMMAND COUNT
---

## Introduction

The `COMMAND COUNT` command retrieves the total number of commands supported by the DiceDB server. It returns an integer value representing the current command count, offering insight into the size of the command set available for use.  
**Time Complexity**: O(1)

## Syntax

```
COMMAND COUNT
```

## Parameters

This command does not accept any parameters.

## Return values

- **Integer**: The command returns an integer representing the total number of commands available in the DiceDB server.
  - For example, if there are 87 commands, the return value will be `(integer) 87`.

### Output format

```
(integer) number_of_commands
```

## Behavior

When executed, the `COMMAND COUNT` command scans the command registry of the DiceDB server and counts the number of registered commands. This allows users to determine the current command count quickly. The operation is efficient and performed in constant time, as the server maintains this information internally.

## Errors

- **Error: unknown sucommand**: This error may occur if the subcommand is misspelled or not recognized by the server.
- `(error) ERR unknown subcommand 'sucommand-name'. Try COMMAND HELP.`

## Example

```bash
127.0.0.1:7379> COMMAND COUNT
(integer) 117
```
