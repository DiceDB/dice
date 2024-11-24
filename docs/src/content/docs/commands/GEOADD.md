---
title: GEOADD
description: The `GEOADD` command in DiceDB is used to add geospatial items (longitude,latitude) to a specified key, storing them as a sorted set. This would allow for efficient querying of geographical data using commands like GEOSEARCH.
---

The `GEOADD` command in DiceDB is used to add geospatial items (longitude,latitude) to a specified key, storing them as a sorted set. This would allow for efficient querying of geographical data using commands like GEOSEARCH.

## Syntax

```bash
GEOADD key [NX | XX] [CH] longitude latitude member [longitude latitude member ...]
```

## Parameters

| Parameter | Description                                                                      | Type   | Required |
| --------- | -------------------------------------------------------------------------------- | ------ | -------- |
| key       | The name of the sorted set where the geospatial data will be stored.             | string | Yes      |
| NX        | Only add new elements; do not update existing ones.                              | NONE   | No       |
| XX        | Only update existing elements; do not add new ones.                              | NONE   | No       |
| longitude | longitude of the location (must be between -180 and 180 degrees).                | float  | Yes      |
| latitude  | latitude of the location (must be between -85.05112878 and 85.05112878 degrees). | float  | Yes      |
| member    | A unique identifier for the location.                                            | string | Yes      |

## Return Values

| Condition                                                  | Return Value                                         |
| ---------------------------------------------------------- | ---------------------------------------------------- |
| For each new member added.                                 | 1                                                    |
| No new member is added.                                    | 0                                                    |
| Incorrect Argument Count                                   | `ERR wrong number of arguments for 'geoadd' command` |
| If the longitude is not a valid number or is out of range. | `ERR invalid longitude`                              |
| If the latitude is not a valid number or is out of range.  | `ERR invalid latitude`                               |

## Behaviour

When the GEOADD command is issued, DiceDB performs the following steps:

1. It checks whether longitude and latitude are valid or not. If not an error is thrown.
2. It checks whether the set exists or not.
3. If set doesn't exist new set is created or else the same set is used.
4. It adds or updates the member in the set.
5. It returns number of members added.

## Errors

1.`Wrong number of arguments for 'GEOADD' command`

- Error Message: (error) ERR wrong number of arguments for 'geoadd' command.
- Occurs when the command is executed with an incorrect number of arguments.

2. `Longitude not a valid number or is out of range `

   - Error Message: (error) ERR invalid longitude.
   - Occurs when longitude is out of range(-180 to 180) or not a valid number.

3. `Latitude not a valid number or is out of range `
   - Error Message: (error) ERR invalid latitude.
   - Occurs when latitude is out of range(-85.05112878 to 85.05112878) or not a valid number.

## Example Usage

Here are a few examples demonstrating the usage of the GEOADD command:

### Example : Adding new member to a set

```bash
127.0.0.1:7379> GEOADD locations 13.361389 38.115556 "Palermo"
1
```

### Example : Updating an already existing member to a set

```bash
127.0.0.1:7379> GEOADD locations 13.361389 39.115556 "Palermo"
0
```

### Example : Error Adding a member with invalid longitude

```bash
127.0.0.1:7379> GEOADD locations 181.120332 39.115556 "Jamaica"
(error) ERROR invalid longitude
```

### Example : Error Adding a member with invalid latitde

```bash
127.0.0.1:7379> GEOADD locations 13.361389 91.115556 "Venice"
(error) ERROR invalid latitude
```
