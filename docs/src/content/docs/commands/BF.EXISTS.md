---
title: BF.EXISTS
description: Documentation for the DiceDB command BF.EXISTS
---

A Bloom Filter is a probabilistic data structure that is used to test whether an element is a member of a set. It is highly space-efficient but allows for a small probability of false positives. The `BF.EXISTS` command checks whether a specified item may exist in the Bloom Filter.

## Syntax

```bash
BF.EXISTS key item
```

## Parameters

| Parameter | Description                                          | Type   | Required |
| --------- | ---------------------------------------------------- | ------ | -------- |
| `key`     | The key under which the Bloom Filter is stored.      | String | Yes      |
| `item`    | The item to check for existence in the Bloom Filter. | String | Yes      |

## Return Value

| Condition                                          | Return Value |
| -------------------------------------------------- | ------------ |
| Item may exist in the Bloom Filter                 | `1`          |
| Item definitely does not exist in the Bloom Filter | `0`          |
| Key does not exist                                 | `0`          |

## Behaviour

When the `BF.EXISTS` command is executed, it checks the Bloom Filter associated with the specified key to determine if the given item may be present. Due to the nature of Bloom Filters, the command can return false positives but never false negatives. This means:

- If the command returns `1`, the item may exist in the Bloom Filter.
- If the command returns `0`, the item definitely does not exist in the Bloom Filter.

## Errors

The `BF.EXISTS` command can raise errors in the following scenarios:

1. `Incorrect number of arguments`:
   - Error message: `(error) ERR wrong number of arguments for 'bf.exists' command`
   - The command requires exactly two arguments: the key and the item to check for existence in the Bloom Filter.
2. `Key is not a Bloom Filter`:
   - Error message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - The specified key does not refer to a Bloom Filter.

## Example Usage

### Checking for an existing item

```bash
127.0.0.1:7379> BF.ADD myBloomFilter "apple"
(integer) 1
127.0.0.1:7379> BF.EXISTS myBloomFilter "apple"
(integer) 1
```

In this example, the item "apple" is added to the Bloom Filter `myBloomFilter`. When we check for the existence of "apple", the command returns `1`, indicating that the item may exist in the Bloom Filter.

### Checking for a non-existing item

```bash
127.0.0.1:7379> BF.EXISTS myBloomFilter "banana"
(integer) 0
```

In this example, the item "banana" is checked in the Bloom Filter `myBloomFilter`. The command returns `0`, indicating that the item definitely does not exist in the Bloom Filter.

### Handling a non-existing key

```bash
127.0.0.1:7379> BF.EXISTS nonExistingKey "apple"
(integer) 0
```

In this example, the key `nonExistingKey` does not exist in the database. The command returns `0`, indicating that the item "apple" definitely does not exist in the Bloom Filter.

### Handling a wrong type of key

```bash
127.0.0.1:7379> SET myString "hello"
OK
127.0.0.1:7379> BF.EXISTS myString "apple"
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

In this example, the key `myString` is associated with a string value, not a Bloom Filter. The command raises a `WRONGTYPE` error.

### Incorrect number of arguments

```bash
127.0.0.1:7379> BF.EXISTS myBloomFilter
(error) ERR wrong number of arguments for 'bf.exists' command
```

In this example, the command is called with an incorrect number of arguments. The command raises an `ERR wrong number of arguments` error.
