---
title: BF.RESERVE
description: Documentation for the DiceDB command BF.RESERVE
---

A Bloom Filter is a probabilistic data structure that is used to test whether an element is a member of a set. It is highly space-efficient but allows for a small probability of false positives. The `BF.RESERVE` command is used to initialize a new Bloom Filter.

## Syntax

```bash
BF.RESERVE key [options]
```

## Parameters

| Parameter          | Description                                                                   | Type    | Required |
| ------------------ | ----------------------------------------------------------------------------- | ------- | -------- |
| `key`              | The key under which the Bloom Filter will be stored.                          | String  | Yes      |
| `error_rate`       | The desired probability of false positives. The default value is 0.01 (1%).   | Float   | No       |
| `initial_capacity` | The initial capacity of the Bloom Filter. The default value is 1000 elements. | Integer | No       |

## Return Value

| Condition            | Return Value              |
| -------------------- | ------------------------- |
| Command successful   | Simple string reply: `OK` |
| Unsuccessful command | `error`                   |

## Behaviour

When the `BF.RESERVE` command is executed, DiceDB will create a new Bloom Filter with the specified key and options. If the key already exists and is associated with a different data type, an error will be raised. The Bloom Filter will be initialized with the specified error rate and initial capacity, which will determine its size and performance characteristics.

## Errors

1. `Incorrect Number of Arguments`:

- Error message: `(error) ERR wrong number of arguments for 'BF.RESERVE' command`
- This error occurs when the command does not have the correct number of arguments.

2. `Invalid Error Rate Range`:

- Error message: `(error) Err (0 < error rate range < 1) `
- This error occurs when the error rate is not within the valid range (0, 1).

3. `Invalid Initial Capacity`:

- Error message: `(error) ERR (capacity should be larger than 0)`
- This error occurs when the initial capacity is not a valid positive integer.

4. `Invalid Data Type`:

- Error message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
- This error occurs when the key is already associated with a value of a different data type.

5. `Bad error rate`:

- Error message: `(error) ERR bad error rate`
- This error occurs when the error rate is not a valid float value.

6. `Bad initial capacity`:

- Error message: `(error) ERR bad capacity`
- This error occurs when the initial capacity is not a valid integer value.

## Example Usage

### Basic Usage

```bash
127.0.0.1:7379> BF.RESERVE my_bloom_filter 0.005 5000
OK
```

This command initializes a Bloom Filter named `my_bloom_filter` with an error rate of 0.005 (0.5%) and an initial capacity of 5000 elements.

### Wrong Data Error

```bash
127.0.0.1:7379> SET my_bloom_filter "value"
OK
127.0.0.1:7379> BF.RESERVE my_bloom_filter 0.01 1000
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

This sequence of commands will result in an error because `my_bloom_filter` is already associated with a string value.

### Invalid Error Rate Range

```bash
127.0.0.1:7379> BF.RESERVE my_bloom_filter error_rate -1.5
(error) Err (0 < error rate range < 1)
```

### Invalid Initial Capacity

```bash
127.0.0.1:7379> BF.RESERVE my_bloom_filter 0.01 initial_capacity -100
(error) ERR (capacity should be larger than 0)
```

This command will result in an error because the initial capacity is not a valid positive integer.
