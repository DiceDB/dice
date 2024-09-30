---
title: GETDEL
description: Documentation for the DiceDB command GETDEL
---

The `GETDEL` command in DiceDB is used to retrieve the value of a specified key and then delete the key from the database. This command is useful when you need to fetch a value and ensure that it is removed from the database in a single atomic operation.

## Syntax

```plaintext
GETDEL key
```

## Parameters

- `key`: The key whose value you want to retrieve and delete. This parameter is a string and must be a valid key in the DiceDB database.

## Return Value

- `String`: If the key exists, the command returns the value associated with the key.
- `nil`: If the key does not exist, the command returns `nil`.

## Behaviour

When the `GETDEL` command is executed, the following steps occur:

1. The command checks if the specified key exists in the DiceDB database.
2. If the key exists, the value associated with the key is retrieved.
3. The key is then deleted from the database.
4. The retrieved value is returned to the client.
5. If the key does not exist, `nil` is returned, and no deletion occurs.

## Error Handling

The `GETDEL` command can raise errors in the following scenarios:

1. `Wrong Type Error`: If the key exists but is not a string (e.g., it is a list, set, hash, etc.), a `WRONGTYPE` error will be raised.
2. `Syntax Error`: If the command is called without the required parameter, a syntax error will be raised.

## Example Usage

### Example 1: Key Exists

```plaintext
SET mykey "Hello, World!"
GETDEL mykey
```

`Output:`

```plaintext
"Hello, World!"
```

`Explanation:`

- The key `mykey` is set with the value `"Hello, World!"`.
- The `GETDEL` command retrieves the value `"Hello, World!"` and deletes the key `mykey` from the database.

### Example 2: Key Does Not Exist

```plaintext
GETDEL nonexistingkey
```

`Output:`

```plaintext
(nil)
```

`Explanation:`

- The key `nonexistingkey` does not exist in the database.
- The `GETDEL` command returns `nil` since the key is not found.

### Example 3: Key of Wrong Type

```plaintext
LPUSH mylist "item1"
GETDEL mylist
```

`Output:`

```plaintext
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

`Explanation:`

- The key `mylist` is a list, not a string.
- The `GETDEL` command raises a `WRONGTYPE` error because it expects the key to be a string.
