---
title: BF.INFO
description: Documentation for the DiceDB command BF.INFO
---
<!-- description -->
BF.INFO command is used to get the information about the Bloom Filter key. 

## Syntax

```bash
BF.INFO key [CAPACITY | SIZE | FILTERS | ITEMS | EXPANSION]
```

## Parameters
| Parameter | Description                                                               | Type    | Required |
|-----------|---------------------------------------------------------------------------|---------|----------|
| `key`       | The name of the Bloom Filter key to get information about.                | string  | Yes      |
| `CAPACITY`  | Optional. Returns the capacity of the Bloom Filter.                       | string  | No       |
| `SIZE`      | Optional. Returns the size of the Bloom Filter.                           | string  | No       |
| `FILTERS`   | Optional. Returns the number of filters in the Bloom Filter.              | string  | No       |
| `ITEMS`     | Optional. Returns the number of items added to the Bloom Filter.          | string  | No       |
| `EXPANSION` | Optional. Returns the expansion factor of the Bloom Filter.               | string  | No       |

Only one of the optional parameters can be used at a time.



## Return values

| Condition                                      | Return Value                                      |
|------------------------------------------------|---------------------------------------------------|
| Command executed successfully                  | Information about the Bloom Filter as an array with capacity, size, filters, items, expansion rate             |
| If an optional parameter is used               | Information about the specific parameter         |
| Error such as invalid number of arguments, invalid parameters      | `(error)`                                   |,

## Behaviour

- When the `BF.INFO` command is executed, it returns information about the Bloom Filter key specified.
- The command can return the capacity, size, number of filters, number of items, and expansion factor of the Bloom Filter.
- If an optional parameter is used, the command returns information about the specific parameter.

## Errors

1. `Wrong type of value or key`:
   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when attempting to use the command on a key that contains a non-string value.
2. `Invalid number of arguments`:
   - Error Message: `(error) ERR wrong number of arguments for 'bf.info' command`
   - Occurs when the command is not provided with the correct number of arguments.
3. `Invalid parameter`:
    - Error Message: `(error) ERR invalid parameter`
    - Occurs when an invalid parameter is provided with the command.
4. `Key does not exist`:
    - Error Message: `(error) ERR not found`
    - Occurs when the specified key does not exist in the database.

## Example Usage

### Basic Usage

```bash
127.0.0.1:7379> BF.INFO myBloomFilter
1) Capacity
2) 100
3) Size
4) 50
5) Filters
6) 1
7) Items
8) 10
9) Expansion
10) 2
```
### Getting Capacity
<!-- getting capacity -->
```bash
127.0.0.1:7379> BF.INFO myBloomFilter CAPACITY
1) Capacity
2) 100
```

### Invalid parameter

```bash
127.0.0.1:7379> BF.INFO myBloomFilter INVALID
(error) ERR invalid parameter
```

### Key does not exist

```bash
127.0.0.1:7379> BF.INFO nonExistingKey
(error) ERR not found
```

### Wrong type of value or key

```bash
127.0.0.1:7379> SET myString "hello"
OK
127.0.0.1:7379> BF.INFO myString
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

## Notes

1. The `CAPACITY` parameter returns the capacity of the Bloom Filter.
    - The capacity is the maximum number of elements that can be added to the Bloom Filter.
2. The `SIZE` parameter returns the size of the Bloom Filter.
    - The size is the number of bits used in the Bloom Filter.
3. The `FILTERS` parameter returns the number of filters in the Bloom Filter.
    - The number of filters is the number of hash functions used in the Bloom Filter.
    - The number of filters is also the number of bits set to 1 when an element is added to the Bloom Filter.
    - More filters reduce the false positive rate but increase the memory usage.
4. The `ITEMS` parameter returns the number of items added to the Bloom Filter.
    - The number of items is the total number of elements added to the Bloom Filter.
5. The `EXPANSION` parameter returns the expansion factor of the Bloom Filter.
    - The expansion factor is the ratio of the number of bits in the Bloom Filter to the number of elements added to the Bloom Filter.
    - A higher expansion factor indicates a higher false positive rate.

