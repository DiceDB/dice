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

- `key`: The key of the set from which the members will be removed. This key must be of the set data type.
- `member`: One or more members to be removed from the set. Multiple members can be specified, separated by spaces.

## Return Value

- `Integer reply`: The number of members that were removed from the set, not including non-existing members.

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

### Example 1: Removing a single member from a set

```bash
SADD myset "one" "two" "three"
SREM myset "two"
```

`Expected Output:`

```
(integer) 1
```

`Explanation:` The member "two" is removed from the set `myset`. The command returns 1 because one member was removed.

### Example 2: Removing multiple members from a set

```bash
SADD myset "one" "two" "three"
SREM myset "two" "three"
```

`Expected Output:`

```
(integer) 2
```

`Explanation:` The members "two" and "three" are removed from the set `myset`. The command returns 2 because two members were removed.

### Example 3: Removing a non-existing member from a set

```bash
SADD myset "one" "two" "three"
SREM myset "four"
```

`Expected Output:`

```
(integer) 0
```

`Explanation:` The member "four" does not exist in the set `myset`. The command returns 0 because no members were removed.

### Example 4: Removing members from a non-existing set

```bash
SREM myset "one"
```

`Expected Output:`

```
(integer) 0
```

`Explanation:` The set `myset` does not exist. The command returns 0 because no members were removed.

### Example 5: Error when key is not a set

```bash
SET mykey "value"
SREM mykey "one"
```

`Expected Output:`

```
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

`Explanation:` The key `mykey` exists but is not of the set data type. The command returns an error.

## Notes

- The `SREM` command is idempotent. Removing a member that does not exist in the set has no effect and does not produce an error.
- The command can be used to remove multiple members in a single call, which can be more efficient than calling `SREM` multiple times for individual members.
- If the set becomes empty after the removal of members, the key is automatically deleted from the database.

By understanding the `SREM` command, you can effectively manage the members of sets in your DiceDB database, ensuring efficient and error-free operations.

