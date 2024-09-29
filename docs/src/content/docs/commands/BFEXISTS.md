---
title: BFEXISTS
description: Documentation for the DiceDB command BFEXISTS
---

A Bloom Filter is a probabilistic data structure that is used to test whether an element is a member of a set. It is highly space-efficient but allows for a small probability of false positives. The `BFEXISTS` command checks whether a specified item may exist in the Bloom Filter.

## Syntax

```plaintext
BFEXISTS key item
```

## Parameters

- `key`: The key under which the Bloom Filter is stored. This is a string.
- `item`: The item to check for existence in the Bloom Filter. This is a string.

## Return Value

- `Integer reply`: The command returns `1` if the item may exist in the Bloom Filter, and `0` if the item definitely does not exist.

## Behaviour

When the `BFEXISTS` command is executed, it checks the Bloom Filter associated with the specified key to determine if the given item may be present. Due to the nature of Bloom Filters, the command can return false positives but never false negatives. This means:

- If the command returns `1`, the item may exist in the Bloom Filter.
- If the command returns `0`, the item definitely does not exist in the Bloom Filter.

## Error Handling

The `BFEXISTS` command can raise errors in the following scenarios:

1. `Key does not exist`: If the specified key does not exist in the database, the command will return `0` without raising an error.
2. `Wrong type of key`: If the key exists but is not associated with a Bloom Filter, a `WRONGTYPE` error will be raised.
3. `Incorrect number of arguments`: If the command is called with an incorrect number of arguments, a `ERR wrong number of arguments` error will be raised.

## Example Usage

### Example 1: Checking for an existing item

```plaintext
127.0.0.1:6379> BFADD myBloomFilter "apple"
(integer) 1
127.0.0.1:6379> BFEXISTS myBloomFilter "apple"
(integer) 1
```

In this example, the item "apple" is added to the Bloom Filter `myBloomFilter`. When we check for the existence of "apple", the command returns `1`, indicating that the item may exist in the Bloom Filter.

### Example 2: Checking for a non-existing item

```plaintext
127.0.0.1:6379> BFEXISTS myBloomFilter "banana"
(integer) 0
```

In this example, the item "banana" is checked in the Bloom Filter `myBloomFilter`. The command returns `0`, indicating that the item definitely does not exist in the Bloom Filter.

### Example 3: Handling a non-existing key

```plaintext
127.0.0.1:6379> BFEXISTS nonExistingKey "apple"
(integer) 0
```

In this example, the key `nonExistingKey` does not exist in the database. The command returns `0`, indicating that the item "apple" definitely does not exist in the Bloom Filter.

### Example 4: Handling a wrong type of key

```plaintext
127.0.0.1:6379> SET myString "hello"
OK
127.0.0.1:6379> BFEXISTS myString "apple"
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

In this example, the key `myString` is associated with a string value, not a Bloom Filter. The command raises a `WRONGTYPE` error.

### Example 5: Incorrect number of arguments

```plaintext
127.0.0.1:6379> BFEXISTS myBloomFilter
(error) ERR wrong number of arguments for 'BFEXISTS' command
```

In this example, the command is called with an incorrect number of arguments. The command raises an `ERR wrong number of arguments` error.
