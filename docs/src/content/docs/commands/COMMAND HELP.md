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

## Return values

| Condition             | Return Value                                                             |
| --------------------- | ------------------------------------------------------------------------ |
| Command is successful | Help text detailing the available subcommands for the `COMMAND` command. |
| Error                 | An error is returned if the command fails.                               |

## Behavior

The `COMMAND HELP` command outputs help text that lists all the available subcommands for the [`COMMAND`](/commands/command) command. It is used to understand how each subcommand functions and the available options for those subcommands.

## Errors

1.  `Arity Error`
    - Error Message: `(error) ERR wrong number of arguments for 'command|help' command`
    - Occurs when invalid number of arguments provided to `COMMAND HELP` command.

## Example Usage

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

An error is thrown when extra arguments are passed to the `COMMAND HELP` command, as it does not accept any additional arguments.

```bash
127.0.0.1:7379> COMMAND HELP EXTRA ARGS
(error) ERR wrong number of arguments for 'command|help' command
```
