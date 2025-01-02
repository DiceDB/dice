---
title: LINDEX
description: The `LINDEX` command in DiceDB is used to find an element present in the list stored at a key. 
---

# LINDEX

The `LINDEX` command in DiceDB is used to find an element present in the list stored at a key. If the key does not exist or the index specified is out of range of the list then the command will throw an error.

## Command Syntax

```bash
LINDEX [key] [index]
```

## Parameters

| Parameter    | Description                                                           |    Type     | Required |
| ------------ | ----------------------------------------------------------------------------------- | -------- | -------- |
| KEY          | The key associated with the list for which the element you want to retrieve.        | String      | Yes      |
| INDEX        | The index or position of the element we want to retrieve in the list. 0-based indexing is used to consider elements from head or start of the list. Negative indexing is used (starts from -1) to consider elements from tail or end of the list.  | Integer       | Yes      |

## Return Values

| Condition                                     | Return Value                                  |
| --------------------------------------------- | --------------------------------------------- |
| Command is successful                         | Returns the element present in that index     |
| Key does not exist                            | error                                         |
| Syntax or specified constraints are invalid    | error                                         |

## Behaviour

When the `LINDEX` command is executed, it performs the specified subcommand operation -

- Returns element present at the `index` of the list associated with the `key` provided as arguments of the command.

- If the `key` exists but is not associated with the list, an error is returned.

## Errors

- `Non existent key`:

  - Error Message: `ERR could not perform this operation on a key that doesn't exist`

- `Missing Arguments`:

  - Error Message: `ERR wrong number of arguments for 'latency subcommand' command`
  - If required arguments for a subcommand are missing, DiceDB will return an error.

- `Key not holding a list`

    - Error Message : `WRONGTYPE Operation against a key holding the wrong kind of value`

## Example Usage

### Basic Usage

```
dicedb> LPUSH k 1
1
dicedb> LPUSH k 2
2
dicedb> LINDEX k 0
2
dicedb> LINDEX k -1
1
```

### Index out of range

```
dicedb> LINDEX k 3
Error: ERR Index out of range
```

### Non-Existent Key

```
dicedb> LINDEX NON-EXISTENT -1
Error: ERR could not perform this operation on a key that doesn't exist
```

## Best Practices

- Check Key Type: Before using `LINDEX`, ensure that the key is associated with a list to avoid errors.

- Handle Non-Existent Keys: Be prepared to handle the case where the key does not exist, as `LINDEX` will return `ERROR` in such scenarios.

- Make sure the index is within range of the list.