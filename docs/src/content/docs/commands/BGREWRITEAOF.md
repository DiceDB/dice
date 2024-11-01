---
title: BGREWRITEAOF
description: The `BGREWRITEAOF` command in DiceDB is used to asynchronously rewrite the Append-Only File (AOF). This command triggers a background process that creates a new AOF file, which is a more compact and optimized version of the current AOF file. The new AOF file will contain the minimal set of commands needed to reconstruct the current dataset.
---

The `BGREWRITEAOF` command in DiceDB is used to asynchronously rewrite the Append-Only File (AOF). This command triggers a background process that creates a new AOF file, which is a more compact and optimized version of the current AOF file. The new AOF file will contain the minimal set of commands needed to reconstruct the current dataset.

## Syntax

```bash
BGREWRITEAOF
```

## Return values

| Condition                                      | Return Value                                      |
|------------------------------------------------|---------------------------------------------------|
| Command is successful                          | `OK`                                              |
| Syntax or specified constraints are invalid    | error                                             |

## Behaviour

When the `BGREWRITEAOF` command is issued, DiceDB performs the following steps:

1. `Forking a Child Process`: DiceDB forks a child process to handle the AOF rewrite. This ensures that the main DiceDB server can continue to handle client requests without interruption.
2. `Creating a Temporary AOF File`: The child process creates a temporary AOF file and writes the minimal set of commands needed to reconstruct the current dataset.
3. `Swapping Files`: Once the temporary AOF file is fully written and synced to disk, the child process swaps the temporary file with the existing AOF file.
4. `Cleaning Up`: The child process exits, and the main DiceDB server continues to operate with the new, optimized AOF file.

## Errors

1. `Unable to create/write AOF file`:

   - Error Message: `ERR AOF failed`
   - Occurs when diceDB is unable to create or write into the AOF file

2. `Forking Error`:

   - Error Message: `ERR Fork failed`
   - Occurs when diceDB is unable to fork a new process due to system limitations or resource constraints.

## Example Usage

### Basic Usage
```bash
127.0.0.1:7379> BGREWRITEAOF
OK
```
