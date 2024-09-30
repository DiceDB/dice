---
title: SADD
description: Documentation for the DiceDB command SADD
---

The `SADD` command in DiceDB is used to add one or more members to a set stored at a specified key. If the specified key does not exist, a new set is created before adding the members. Members that are already present in the set are ignored. This command is useful for maintaining collections of unique elements.

## Syntax

```plaintext
SADD key member [member ...]
```

## Parameters

- `key`: The key of the set to which the members will be added. If the key does not exist, a new set is created.
- `member`: One or more members to be added to the set. Multiple members can be specified, separated by spaces.

## Return Value

- `Integer`: The number of elements that were added to the set, not including all the elements already present in the set.

## Behaviour

When the `SADD` command is executed, the following actions occur:

1. `Key Existence Check`: DiceDB checks if the key exists.
   - If the key does not exist, a new set is created.
   - If the key exists but is not a set, an error is returned.
2. `Member Addition`: Each specified member is added to the set.
   - If a member is already present in the set, it is ignored.
3. `Return Value Calculation`: The command returns the number of members that were actually added to the set.

## Error Handling

- `WRONGTYPE Operation against a key holding the wrong kind of value`: This error is returned if the key exists and is not a set. DiceDB expects the key to be associated with a set data type. If the key is associated with a different data type (e.g., a string, list, or hash), the command will fail with this error.

## Example Usage

### Example 1: Adding Members to a New Set

```plaintext
SADD myset "member1" "member2" "member3"
```

`Explanation`: This command creates a new set with the key `myset` and adds the members `member1`, `member2`, and `member3` to it. Since the set is new, all three members are added.

`Return Value`: `3` (all three members were added)

### Example 2: Adding Members to an Existing Set

```plaintext
SADD myset "member2" "member4"
```

`Explanation`: This command adds the members `member2` and `member4` to the existing set `myset`. Since `member2` is already present in the set, only `member4` is added.

`Return Value`: `1` (only `member4` was added)

### Example 3: Handling Non-Set Key

```plaintext
SET mykey "value"
SADD mykey "member1"
```

`Explanation`: The first command sets the key `mykey` to a string value `"value"`. The second command attempts to add `member1` to `mykey`, but since `mykey` is not a set, an error is returned.

`Error`: `WRONGTYPE Operation against a key holding the wrong kind of value`

## Additional Notes

- The `SADD` command is idempotent with respect to the members already present in the set. Adding the same member multiple times will not change the set or the return value.
- The command can handle multiple members in a single call, which can be more efficient than adding members one by one.
- The order of members in a set is not guaranteed, as sets are unordered collections.

By understanding the `SADD` command, you can effectively manage sets in DiceDB, ensuring that your collections of unique elements are maintained efficiently and correctly.

