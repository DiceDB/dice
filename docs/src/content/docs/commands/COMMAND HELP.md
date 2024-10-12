---
title: COMMAND HELP
description: Documentation for the DiceDB command COMMAND HELP
---

## Introduction

The `COMMAND HELP` command provides details on all the available subcommands for the `COMMAND` command in DiceDB. This command is useful for obtaining information on what each subcommand does and how to use them.

## Syntax

```bash
COMMAND HELP
```

## Parameters

This command does not accept any parameters.

## Return Values

Returns an array of help text entries. Each entry describes a subcommand, its syntax, and its purpose.

## Behavior

The `COMMAND HELP` command outputs help text that lists all the available subcommands for the `COMMAND` command. It is used to understand how each subcommand functions and the available options for those subcommands.

## Errors

- **Error: Arity Error**: Returned when invalid number of arguments provided.
  ```bash
  (error) ERR wrong number of arguments for 'command|help' command
  ```

## Examples

### Print help

```bash
127.0.0.1:7379> COMMAND HELP
 1) "COMMAND <subcommand> [<arg> [value] [opt] ...]. Subcommands are:"
 2) "(no subcommand)"
 3) "     Return details about all DiceDB commands."
 4) "COUNT"
 5) "     Return the total number of commands in this DiceDB server."
 6) "LIST"
 7) "     Return a list of all commands in this DiceDB server."
 8) "INFO [<command-name> ...]"
 9) "     Returns details about the specified DiceDB commands. If no command names are given, documentation details for all commands are returned"
 10) "GETKEYS <full-command>"
 11) "     Return the keys from a full DiceDB command."
 12) "HELP"
 13) "     Print this help."
```

### Arity Error

```bash
127.0.0.1:7379> COMMAND HELP EXTRA ARGS
(error) ERR wrong number of arguments for 'command|help' command
```
