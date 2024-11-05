---
title: HMGET
description: The `HMGET` command in DiceDB is used to retrieve the values of one or more specified fields from a hash. It allows efficient fetching of specific fields from a hash without retrieving the entire hash.
---

The `HMGET` command in DiceDB is used to retrieve the values of one or more specified fields from a hash. It allows efficient fetching of specific fields from a hash without retrieving the entire hash.

## Syntax

```bash
HMGET key field [field ...]
```

## Parameters

| Parameter       | Description                                                   | Type    | Required |
|-----------------|---------------------------------------------------------------|---------|----------|
| `key`           | The name of the hash.                                         | String  | Yes      |
| `field`         | The field within the hash to retrieve the value for.          | String  | Yes      |
| `[field ...]`   | Additional fields to retrieve from the hash.                  | String  | No       |

## Return Values

| Condition                                    | Return Value                                                                |
|----------------------------------------------|-----------------------------------------------------------------------------|
| Field exists                                 | `String` (The value of the field)                                           |
| Field does not exist                         | `nil`                                                                       |
| Multiple fields retrieved                    | List of values for each field                                               |
| Wrong data type                              | `(error) WRONGTYPE Operation against a key holding the wrong kind of value` |
| Incorrect Argument Count                     | `(error) ERR wrong number of arguments for 'hmget' command`                 |

## Behaviour

When the `HMGET` command is executed, the following actions occur:

- The specified fields are fetched from the hash.
- If a field exists, its value is returned.
- If a field does not exist, `nil` is returned for that field.
- The command returns a list of values, corresponding to each field requested.

## Errors

The `HMGET` command can raise errors in the following scenarios:

1. `Non-hash type or wrong data type`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs if `key` holds a non-hash data structure.

2. `Incorrect Argument Count`:

   - Error Message: `(error) ERR wrong number of arguments for 'hmget' command`
   - Occurs if the command is not provided with the correct number of arguments (i.e., fewer than two).

## Example Usage

### Basic Usage

#### Retrieving Multiple Fields

```bash
127.0.0.1:7379> HMGET product:2000 name price stock
1) "Laptop"
2) "999.99"
3) "50"
```
- **Behaviour**: The values of the fields `name`, `price`, and `stock` from the hash `product:2000` are retrieved and returned in the order specified.
- **Return Value**: List of field values.

#### Retrieving Fields with Missing Values

```bash
127.0.0.1:7379> HMGET product:2000 name description
1) "Laptop"
2) (nil)
```
- **Behaviour**: The `name` field exists, so its value is returned. The `description` field does not exist in the hash, so `nil` is returned.
- **Return Value**: List with value and `nil`.

### Invalid Usage

Trying to retrieve fields from a key that is not a hash.

```bash
127.0.0.1:7379> SET product:2000 "This is a string"
OK
127.0.0.1:7379> HMGET product:2000 name price
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

- **Behaviour**: The `SET` command sets the key `product:2000` to a string value.
- **Error**: The `HMGET` command will raise a `WRONGTYPE` error because `product:2000` is not a hash.

Missing Key or Field Arguments

```bash
127.0.0.1:7379> HMGET
(error) ERR wrong number of arguments for 'hmget' command

127.0.0.1:7379> HMGET product:2000
(error) ERR wrong number of arguments for 'hmget' command
```
- **Behavior**: The `HGET` command requires at least two arguments: the key and the field name.
- **Error**: The command fails if no key or fields are specified. DiceDB raises an error indicating that the number of arguments is incorrect.

### Best Practices

- Use `HMGET` to fetch only the fields you need from a hash to minimize data transfer and improve performance.
