---
title: COMMAND LIST
description: Documentation for the DiceDB command COMMAND LIST.
---

## Introduction

The `COMMAND LIST` command retrieves a list of all commands supported by the DiceDB server. This allows users to discover available commands for various operations, making it easier to understand the capabilities of the database.

## Syntax

```bash
COMMAND LIST
```

## Parameters

This command does not accept any parameters.

## Return values

The command returns an array of strings, where each string represents a command name available in the DiceDB server. If no commands are available (which is unlikely), an empty array is returned.

## Behavior

When executed, the `COMMAND LIST` command scans the DiceDB server's command registry and compiles a list of command names. The operation is efficient, leveraging the server's internal command registration system to provide results quickly.

## Errors

1.  Arity Error
    - Error Message: `(error) ERR wrong number of arguments for 'command|list' command`
    - Occurs when invalid number of arguments provided for `COMMAND LIST` command.

## Example Usage

### Retrieve the list of available commands on the DiceDB server

```bash
127.0.0.1:7379> COMMAND LIST
  1) "SLEEP"
  2) "SMEMBERS"
  3) "BFINIT"
  4) "FLUSHDB"
  .
  .
  .
127.0.0.1:7379>
```

### Arity Error

An error is thrown when extra arguments are passed to the `COMMAND LIST` command, as it does not accept any additional arguments.

```bash
127.0.0.1:7379> COMMAND LIST EXTRA ARGS
(error) ERR wrong number of arguments for 'command|list' command
```
