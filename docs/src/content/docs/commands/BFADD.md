---
title: BFADD
description: Documentation for the DiceDB command BFADD
---

A Bloom Filter is a probabilistic data structure that is used to test whether an element is a member of a set. It is highly space-efficient but allows for a small probability of false positives. The `BFADD` command is used to add an element to a Bloom Filter.

## Syntax

```plaintext
BFADD key item
```

## Parameters

- `key`: The name of the Bloom Filter to which the item will be added. This is a string.
- `item`: The item to be added to the Bloom Filter. This is a string.

## Return Value

- `Integer reply`: The command returns `1` if the item was not already present in the Bloom Filter and `0` if the item was already present.

## Behaviour

When the `BFADD` command is executed, the specified item is added to the Bloom Filter associated with the given key. If the Bloom Filter does not already exist, it will be created automatically. The command will then check if the item is already present in the Bloom Filter:

- If the item is not present, it will be added, and the command will return `1`.
- If the item is already present, the command will return `0`.

## Error Handling

The `BFADD` command can raise errors in the following scenarios:

1. `Wrong number of arguments`: If the command is called with an incorrect number of arguments, a `ERR wrong number of arguments for 'BFADD' command` error will be raised.
2. `Non-string key or item`: If the key or item is not a string, a `WRONGTYPE Operation against a key holding the wrong kind of value` error will be raised.

## Example Usage

### Adding an Item to a Bloom Filter

```plaintext
127.0.0.1:6379> BFADD mybloomfilter "apple"
(integer) 1
```

In this example, the item "apple" is added to the Bloom Filter named `mybloomfilter`. Since "apple" was not already present, the command returns `1`.

### Adding an Existing Item to a Bloom Filter

```plaintext
127.0.0.1:6379> BFADD mybloomfilter "apple"
(integer) 0
```

In this example, the item "apple" is added to the Bloom Filter named `mybloomfilter` again. Since "apple" was already present, the command returns `0`.

### Error Scenario: Wrong Number of Arguments

```plaintext
127.0.0.1:6379> BFADD mybloomfilter
(error) ERR wrong number of arguments for 'BFADD' command
```

In this example, the command is called with only one argument instead of the required two. An error is raised indicating the wrong number of arguments.

### Error Scenario: Non-string Key or Item

```plaintext
127.0.0.1:6379> BFADD 12345 67890
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

In this example, the key and item are non-string values. An error is raised indicating the wrong type of value.
