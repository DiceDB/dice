---
title: BFINIT
description: Documentation for the DiceDB command BFINIT
---

A Bloom Filter is a probabilistic data structure that is used to test whether an element is a member of a set. It is highly space-efficient but allows for a small probability of false positives. The `BFINIT` command is used to initialize a new Bloom Filter.

## Syntax

```plaintext
BFINIT key [options]
```

## Parameters

- `key`: (Required) The name of the Bloom Filter to be created. This key will be used to reference the Bloom Filter in subsequent operations.
- `options`: (Optional) A set of options to configure the Bloom Filter. These options can include:
  - `error_rate`: The desired false positive rate. This is a floating-point number between 0 and 1. The default value is 0.01 (1%).
  - `initial_capacity`: The initial number of elements that the Bloom Filter is expected to hold. The default value is 1000.

## Return Value

- `Simple String Reply`: Returns `OK` if the Bloom Filter is successfully created.

## Behaviour

When the `BFINIT` command is executed, DiceDB will create a new Bloom Filter with the specified key and options. If the key already exists and is associated with a different data type, an error will be raised. The Bloom Filter will be initialized with the specified error rate and initial capacity, which will determine its size and performance characteristics.

## Error Handling

- `WRONGTYPE`: If the key already exists and is not a Bloom Filter, DiceDB will return an error with the message `WRONGTYPE Operation against a key holding the wrong kind of value`.
- `ERR invalid error rate`: If the provided error rate is not a valid floating-point number between 0 and 1, DiceDB will return an error with the message `ERR invalid error rate`.
- `ERR invalid initial capacity`: If the provided initial capacity is not a valid positive integer, DiceDB will return an error with the message `ERR invalid initial capacity`.

## Example Usage

### Basic Usage

```plaintext
BFINIT my_bloom_filter
```

This command initializes a Bloom Filter named `my_bloom_filter` with the default error rate of 0.01 and an initial capacity of 1000 elements.

### Custom Error Rate and Initial Capacity

```plaintext
BFINIT my_bloom_filter error_rate 0.005 initial_capacity 5000
```

This command initializes a Bloom Filter named `my_bloom_filter` with an error rate of 0.005 (0.5%) and an initial capacity of 5000 elements.

### Error Handling Example

#### WRONGTYPE Error

```plaintext
SET my_bloom_filter "some string value"
BFINIT my_bloom_filter
```

This sequence of commands will result in an error because `my_bloom_filter` is already associated with a string value. The error message will be:

```plaintext
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

#### Invalid Error Rate

```plaintext
BFINIT my_bloom_filter error_rate -0.01
```

This command will result in an error because the error rate is not within the valid range. The error message will be:

```plaintext
(error) ERR invalid error rate
```

#### Invalid Initial Capacity

```plaintext
BFINIT my_bloom_filter initial_capacity -100
```

This command will result in an error because the initial capacity is not a valid positive integer. The error message will be:

```plaintext
(error) ERR invalid initial capacity
```
