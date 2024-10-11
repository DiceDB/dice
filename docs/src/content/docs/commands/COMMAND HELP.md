---
title: COMMAND HELP
description: Documentation for the DiceDB command COMMAND HELP
---

## Introduction

The `COMMAND HELP` command provides details on all the available subcommands for the `COMMAND` command in DiceDB. This command is useful for obtaining information on what each subcommand does and how to use them.

## Syntax

```
COMMAND HELP
```

## Parameters

This command does not accept any parameters.

## Return Values

Returns an array of help text entries. Each entry describes a subcommand, its syntax, and its purpose.

## Behavior

The `COMMAND HELP` command outputs help text that lists all the available subcommands for the `COMMAND` command. It is used to understand how each subcommand functions and the available options for those subcommands.

## Errors

- No specific errors are thrown by `COMMAND HELP` as it does not take any arguments.

## Example

```bash
127.0.0.1:7379> COMMAND HELP
 1) "COMMAND <subcommand> [<arg> [value] [opt] ...]. Subcommands are:"
 2) "(no subcommand)"
 3) "    Return details about all Dice commands."
 4) "COUNT"
 5) "    Return the total number of commands in this Dice server."
 6) "LIST"
 7) "     Return a list of all commands in this Dice server."
 8) "INFO [<command-name> ...]"
 9) "    Returns details about the specified DiceDB commands. If no command names are given, documentation details for all commands are returned"
 10) "GETKEYS <full-command>"
 11) "     Return the keys from a full Dice command."
 12) "HELP"
 13) "     Print this help."
```
