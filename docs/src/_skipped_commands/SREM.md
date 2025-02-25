---
title: SREM
description: Documentation for the DiceDB command SREM
---

The `SREM` command is used to remove one or more members from a set stored at a specified key. If the specified members are not present in the set, they are ignored. If the key does not exist, it is treated as an empty set and the command returns 0. This command is idempotent, meaning that removing a member that does not exist in the set has no effect.

## Syntax

```bash
SREM key member [member ...]
```

## Parameters

| Parameter | Description                                                                                             | Type   | Required |
| --------- | ------------------------------------------------------------------------------------------------------- | ------ | -------- |
| `key`     | The key of the set from which the members will be removed. This key must be of the set data type.       | String | Yes      |
| `member`  | One or more members to be removed from the set. Multiple members can be specified, separated by spaces. | String | Yes      |

## Return Value

| Condition                           | Return Value                                                                                             |
| ----------------------------------- | -------------------------------------------------------------------------------------------------------- |
| `key` does not exist                | 0                                                                                                        |
| `key` is not a set                  | error                                                                                                    |
| `member` does not exist in the set  | 0                                                                                                        |
| Invalid syntax or no specified keys | error                                                                                                    |
| Members are successfully removed    | Integer reply: The number of members that were removed from the set, not including non-existing members. |

## Behaviour

When the `SREM` command is executed, the following steps occur:

1. DiceDB checks if the key exists.
2. If the key does not exist, it is treated as an empty set, and the command returns 0.
3. If the key exists but is not of the set data type, an error is returned.
4. DiceDB attempts to remove the specified members from the set.
5. The command returns the number of members that were successfully removed.

## Error Handling

- `WRONGTYPE Operation against a key holding the wrong kind of value`: This error is returned if the key exists but is not of the set data type.
- `Syntax error`: This error is returned if the command is not used with the correct number of arguments.

## Example Usage

### Removing a single member from a set

```bash
127.0.0.1:7379> SADD myset "one" "two" "three"
127.0.0.1:7379> SREM myset "two"
(integer) 1
```

The member "two" is removed from the set `myset`. The command returns 1 because one member was removed.

### Removing multiple members from a set

```bash
127.0.0.1:7379> SADD myset "one" "two" "three"
127.0.0.1:7379> SREM myset "two" "three"
(integer) 2
```

The members "two" and "three" are removed from the set `myset`. The command returns 2 because two members were removed.

### Removing members from a non-existing set

```bash
127.0.0.1:7379> SREM myset "one"
(integer) 0
```

The set `myset` does not exist. The command returns 0 because no members were removed.

### Error when key is not a set

```bash
127.0.0.1:7379> SET mykey "value"
127.0.0.1:7379> SREM mykey "one"
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

The key `mykey` exists but is not of the set data type. The command returns an error.

## Notes

- The `SREM` command is idempotent. Removing a member that does not exist in the set has no effect and does not produce an error.
- The command can be used to remove multiple members in a single call, which can be more efficient than calling `SREM` multiple times for individual members.
- If the set becomes empty after the removal of members, the key is automatically deleted from the database.

By understanding the `SREM` command, you can effectively manage the members of sets in your DiceDB database, ensuring efficient and error-free operations.
