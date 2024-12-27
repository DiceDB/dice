---
title: BITFIELD
description: The BITFIELD command in DiceDB performs complex bitwise operations on string values at specified offsets, treating them as an array of integers. It allows manipulation of individual bits or groups of bits, supporting commands like SET, GET, and INCRBY to update or retrieve bitfield values.
---

## Syntax

```bash
BITFIELD key [GET type offset | [OVERFLOW <WRAP | SAT | FAIL>]
  <SET type offset value | INCRBY type offset increment>
  [GET type offset | [OVERFLOW <WRAP | SAT | FAIL>]
  <SET type offset value | INCRBY type offset increment>
  ...]]
```

## Parameters

| Parameter                      | Description                                                                                                               | Type   | Required |
| ------------------------------ | ------------------------------------------------------------------------------------------------------------------------- | ------ | -------- |
| `key`                          | The name of the key containing the bitfield.                                                                              | String | Yes      |
| `GET type offset`              | Retrieves bits starting at the specified offset with the specified type. Type defines the signed/unsigned integer format. | String | Optional |
| `SET type offset value`        | Sets bits at the specified offset to the given value using the specified integer type.                                    | String | Optional |
| `INCRBY type offset increment` | Increments bits at the specified offset by the increment value and wraps around on overflow based on type.                | String | Optional |
| `OVERFLOW WRAP\|SAT\|FAIL`     | Defines overflow behavior.                                                                                                | String | Optional |

## Return values

| Condition                                   | Return Value                                      |
| ------------------------------------------- | ------------------------------------------------- |
| Command is successful                       | Array of results corresponding to each subcommand |
| Syntax or specified constraints are invalid | error                                             |

## Behaviour

The BITFIELD command in DiceDB allows for direct bitwise manipulation within a binary string stored in a single key. It works by treating the string as an array of integers and performs operations on specific bits or groups of bits at specified offsets:

- SET: Sets the value of bits at a specific offset.
- GET: Retrieves the value of bits at a specific offset.
- INCRBY: Increments the value at a specific offset by a given amount, useful for counters.
- OVERFLOW: Defines the overflow behavior (WRAP, SAT, FAIL) for INCRBY, determining how to handle values that exceed the bitfield’s limits.

### Overflow Options:

- WRAP: Cycles values back to zero on overflow (default behavior).
- SAT: Saturates to the maximum or minimum value for the bit width.
- FAIL: Prevents overflow by causing INCRBY to fail if it would exceed the limits.
  <br>
  <br>
- Supports unsigned (u) with bit widths between 1 and 63 and signed (i) integers with bit widths between 1 and 64.
- The offset specifies where the bitfield starts within the key’s binary string.
- If an offset is out of range (far beyond the current size), DiceDB will automatically expand the binary string to accommodate it, which can impact memory usage.

## Errors

1. `Syntax Error`:

   - Error Message: `(error) ERR syntax error`
   - Occurs if the commands syntax is incorrect.

2. `Invalid bitfield type`:

   - Error Message: `(error) ERR Invalid bitfield type. Use something like i16 u8. Note that u64 is not supported but i64 is`
   - Occurs when attempting to use the command on a wrong type.

3. `Non-integer value`:

   - Error Message: `(error) ERR value is not an integer or out of range`
   - Occurs when attempting to use the command on a value that contains a non-integer value.

4. `Invalid OVERLOW type`:

   - Error Message: `(error) ERR Invalid OVERFLOW type specified`
   - Occurs when attempting to use a wrong OVERFLOW type.

5. `Wrong type of value or key`:
   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when attempting to use the command on a key that contains a non-string value.
6. `Invalid bIT offset`:
   - Error Message: `(error) ERR bit offset is not an integer or out of range`
   - Occurs when attempting to use the command with an invalid bit offset.

## Example Usage

### Basic Usage:

```bash
127.0.0.1:7379> BITFIELD mykey INCRBY i5 100 1 GET u4 0
1) "1"
2) "0"
```

### Overflow control:

```bash
127.0.0.1:7379> BITFIELD mykey incrby u2 100 1 OVERFLOW SAT incrby u2 102 1
1) "1"
2) "1"
127.0.0.1:7379> BITFIELD mykey incrby u2 100 1 OVERFLOW SAT incrby u2 102 1
1) "2"
2) "2"
127.0.0.1:7379> BITFIELD mykey incrby u2 100 1 OVERFLOW SAT incrby u2 102 1
1) "3"
2) "3"
127.0.0.1:7379> BITFIELD mykey incrby u2 100 1 OVERFLOW SAT incrby u2 102 1
1) "0"
2) "3"
```

### Example of OVERFLOW FAIL returning nil:

```bash
127.0.0.1:7379> BITFIELD mykey OVERFLOW FAIL incrby u2 102 1
(nil)
```

### Invalid usage:

```bash
127.0.0.1:7379> BITFIELD
(error) ERR wrong number of arguments for 'bitfield' command
```

```bash
127.0.0.1:7379> SADD bits a b c
(integer) 3
127.0.0.1:7379> BITFIELD bits
(error) ERR WRONGTYPE Operation against a key holding the wrong kind of value
```

```bash
127.0.0.1:7379> BITFIELD bits SET u8 0 255 INCRBY u8 0 100 GET u8
(error) ERR syntax error
```

```bash
127.0.0.1:7379> bitfield bits set a8 0 255 incrby u8 0 100 get u8
(error) ERR Invalid bitfield type. Use something like i16 u8. Note that u64 is not supported but i64 is
```

```bash
127.0.0.1:7379> bitfield bits set u8 a 255 incrby u8 0 100 get u8
(error) ERR bit offset is not an integer or out of range
```

```bash
127.0.0.1:7379> bitfield bits set u8 0 255 incrby u8 0 100 overflow wraap
(error) ERR Invalid OVERFLOW type specified
```

```bash
127.0.0.1:7379> bitfield bits set u8 0 incrby u8 0 100 get u8 288
(error) ERR value is not an integer or out of range
```

## Notes

Where an integer encoding is expected, it can be composed by prefixing with i for signed integers and u for unsigned integers with the number of bits of our integer encoding. So for example u8 is an unsigned integer of 8 bits and i16 is a signed integer of 16 bits.

The supported encodings are up to 64 bits for signed integers, and up to 63 bits for unsigned integers. This limitation with unsigned integers is due to the fact that currently RESP is unable to return 64 bit unsigned integers as replies.

## Subcommands

- **subcommand**: Optional. Available subcommands include:
  - `GET` `<type>` `<offset>` : Returns the specified bit field.
  - `SET` `<type>` `<offset>` `<value>` : Set the specified bit field and - returns its old value.
  - `INCRBY` `<type>` `<offset>` `<increment>` : Increments or decrements (if a negative increment is given) the specified bit field and returns the new value.
  - `OVERFLOW` [ `WRAP` | `SAT` | `FAIL` ]
