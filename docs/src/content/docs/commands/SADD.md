---
title: SADD
description: The `SADD` command in DiceDB is used to add one or more members to a set. If the key does not exist, a new set is created before adding the specified members. This command is fundamental for creating and updating sets in DiceDB, allowing efficient management of unique collections of elements.
---

The `SADD` command in DiceDB is used to add one or more members to a set. If the key does not exist, a new set is created before adding the specified members. This command is fundamental for creating and updating sets in DiceDB, allowing efficient management of unique collections of elements. Sets in DiceDB are unordered collections of unique strings, making them ideal for scenarios where uniqueness and membership operations are important.

## Syntax

```bash
SADD key member [member ...]
```

## Parameters

| Parameter | Description                                        | Type   | Required | Multiple |
| --------- | -------------------------------------------------- | ------ | -------- | -------- |
| `key`     | The name of the set to which members will be added | String | Yes      | No       |
| `member`  | One or more members to add to the set              | String | Yes      | Yes      |

- `key`: This is the identifier for the set in DiceDB. If the key doesn't exist, SADD creates a new set. If it exists but is not a set, SADD returns an error.
- `member`: These are the string values to be added to the set. You can specify one or more members in a single SADD command. Each member is treated as a unique string.

## Return values

| Condition                   | Return Value                                |
| --------------------------- | ------------------------------------------- |
| Command is successful       | Integer: Number of members added to the set |
| Key exists but is not a set | Error                                       |

- The integer returned represents the number of members that were actually added to the set, not including elements that were already present.
- If all specified members already exist in the set, the command returns 0.

## Behaviour

1. **Key Existence Check**:

   - If the key does not exist, DiceDB creates a new set and adds the specified members to it.
   - If the key exists and is a set, the command adds the new members to the existing set.
   - If the key exists but is not a set, an error is returned.

2. **Member Addition**:

   - Each specified member is added to the set if it's not already present.
   - If a member already exists in the set, it is ignored, and no error is raised.
   - The order in which members are specified does not affect the set, as sets are unordered.

3. **Uniqueness**:
   - Sets only store unique elements. Attempting to add a duplicate element has no effect on the set.
   - The uniqueness is based on string comparison. For example, "1" and 1 are considered different members.

## Errors

1. `Wrong type of value or key`:
   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when attempting to use the command on a key that contains a non-set value (e.g., string, list, hash).
   - This error ensures type safety in DiceDB operations.

## Example Usage

### Basic Usage: Creating a New Set

Adding members to a new set:

```bash
127.0.0.1:7379> SADD fruits "apple" "banana" "cherry"
(integer) 3
```

Explanation: This creates a new set named "fruits" with three members. The command returns 3, indicating all three members were added successfully.

### Adding to an Existing Set

Adding more members to an existing set:

```bash
127.0.0.1:7379> SADD fruits "banana" "date" "elderberry"
(integer) 2
```

Explanation: This adds "date" and "elderberry" to the existing "fruits" set. "banana" is ignored as it already exists. The command returns 2, indicating two new members were added.

### Adding Duplicate Members

Adding members that already exist in the set:

```bash
127.0.0.1:7379> SADD fruits "apple" "fig" "apple"
(integer) 1
```

Explanation: This adds only "fig" to the set. Both instances of "apple" are ignored as "apple" already exists in the set. The command returns 1, indicating one new member was added.

### Adding Multiple Members at Once

Adding several members in a single command:

```bash
127.0.0.1:7379> SADD vegetables "carrot" "broccoli" "spinach" "carrot"
(integer) 3
```

Explanation: This creates a new set "vegetables" and adds three unique members. Note that "carrot" is only added once. The command returns 3, the number of new members added.

### Error Case: Wrong Type

Attempting to add to a non-set key:

```bash
127.0.0.1:7379> SET mykey "value"
OK
127.0.0.1:7379> SADD mykey "member1"
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

Explanation: This first sets "mykey" as a string, then attempts to use SADD on it. The command fails because "mykey" is not a set.

### Checking Set Contents

While not part of the SADD command, you can use SMEMBERS to verify the contents of a set:

```bash
127.0.0.1:7379> SMEMBERS fruits
1) "apple"
2) "banana"
3) "cherry"
4) "date"
5) "elderberry"
6) "fig"
```
