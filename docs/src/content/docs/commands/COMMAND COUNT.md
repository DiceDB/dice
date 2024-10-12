---
title: COMMAND COUNT
description: Documentation for the DiceDB command COMMAND COUNT
---

## Introduction

The `COMMAND COUNT` command retrieves the total number of commands supported by the DiceDB server. It returns an integer value representing the current command count, offering insight into the size of the command set available for use.  
**Time Complexity**: O(1)

## Syntax

```bash
COMMAND COUNT
```

## Parameters

This command does not accept any parameters.

## Return values

- **Integer**: The command returns an integer representing the total number of commands available in the DiceDB server.

### Output format

```bash
(integer) number_of_commands
```

## Behavior

When executed, the `COMMAND COUNT` command scans the command registry of the DiceDB server and counts the number of registered commands. This allows users to determine the current command count quickly. The operation is efficient and performed in constant time, as the server maintains this information internally.

## Errors

- **Error: Arity Error**: Returned when invalid number of arguments provided.
  ```bash
  (error) ERR wrong number of arguments for 'command|count' command
  ```

## Examples

### Retrieve the number of commands supported by the DiceDB server

```bash
127.0.0.1:7379> COMMAND COUNT
(integer) 117
```

### Arity Error

```bash
127.0.0.1:7379> COMMAND COUNT EXTRA ARGS
(error) ERR wrong number of arguments for 'command|count' command
```
