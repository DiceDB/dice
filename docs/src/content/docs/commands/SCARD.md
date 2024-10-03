---
title: SCARD
description: Documentation for the DiceDB command SCARD
---

The `SCARD` command in DiceDB is used to get the number of members in a set. This command is useful for determining the size of a set stored at a given key.

## Syntax

```plaintext
SCARD key
```

## Parameters

- `key`: The key of the set whose cardinality (number of members) you want to retrieve. The key must be a valid string.

## Return Value

- `Integer`: The number of elements in the set, or 0 if the set does not exist.

## Behaviour

When the `SCARD` command is executed, DiceDB will:

1. Check if the key exists.
2. If the key does not exist, it will return 0.
3. If the key exists but is not a set, an error will be returned.
4. If the key exists and is a set, it will return the number of elements in the set.

## Error Handling

The `SCARD` command can raise the following errors:

- `WRONGTYPE Operation against a key holding the wrong kind of value`: This error occurs if the key exists but is not a set. DiceDB expects the key to be associated with a set data type. If the key is associated with a different data type (e.g., a string, list, hash, or sorted set), this error will be raised.

## Example Usage

### Example 1: Basic Usage

```plaintext
DiceDB> SADD myset "apple"
(integer) 1
DiceDB> SADD myset "banana"
(integer) 1
DiceDB> SADD myset "cherry"
(integer) 1
DiceDB> SCARD myset
(integer) 3
```

In this example, we first add three members to the set `myset`. Then, we use the `SCARD` command to get the number of members in the set, which returns 3.

### Example 2: Non-Existent Key

```plaintext
DiceDB> SCARD nonexistingset
(integer) 0
```

In this example, we attempt to get the cardinality of a set that does not exist. The `SCARD` command returns 0, indicating that the set is empty or does not exist.

### Example 3: Wrong Type Error

```plaintext
DiceDB> SET mystring "hello"
OK
DiceDB> SCARD mystring
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

In this example, we first set a string value to the key `mystring`. When we attempt to use the `SCARD` command on this key, DiceDB returns an error because `mystring` is not a set.
