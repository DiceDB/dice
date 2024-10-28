package dicedbadapter

import (
	"fmt"
	"strings"
)

// flattenArgs flattens a list of arguments into a single string.
// arguments can be a list of strings or a list of lists of strings.
// nesting should be uniform.
func flattenStringArgs(args interface{}) (string, error) {
	if v, ok := args.(string); ok {
		return v, nil
	}
	if _, ok := args.([]interface{}); !ok {
		return "", fmt.Errorf("unexpected argument type: %T", args)
	}
	var result []string

	// Helper function to recursively flatten the arguments
	var flatten func([]interface{}) error
	flatten = func(items []interface{}) error {

		for _, item := range items {
			switch v := item.(type) {
			case string:
				// If the item is a string, add it to the result
				result = append(result, v)
			case int, int64, float64, bool, int16, int32, int8, uint, uint16, uint32, uint64:
				// If the item is a number, add it to the result
				result = append(result, fmt.Sprintf("%v", v))
			case []interface{}:
				// If the item is a nested list, recursively flatten it
				if err := flatten(v); err != nil {
					return err
				}
			default:
				// Return an error if an unexpected type is encountered
				return fmt.Errorf("unexpected argument type: %T", item)
			}
		}
		return nil
	}
	// Start the recursive flattening process
	if err := flatten(args.([]interface{})); err != nil {
		return "", err
	}

	// Join all flattened elements with spaces
	return strings.Join(result, " "), nil
}

// orderAndEncodeArgs encodes arguments based on ArgsOrder for any Redis command.
func orderAndEncodeArgs(command string, args map[string]interface{}) (string, error) {
	argsOrder := DiceCmdAdapters[command].ArgsOrder
	// Start with the command name
	result := command

	for _, arg := range argsOrder {
		switch v := arg.(type) {
		case string:
			// For regular fields like "key" and "value"
			if argVal, ok := args[v]; ok {
				// If it's a flag (e.g., "nx": "nx"), only add the value once
				if argVal == v {
					result += fmt.Sprintf(" %s", v)
				} else if v == "key" || v == "value" {
					result += fmt.Sprintf(" %s", argVal)
				} else {
					result += fmt.Sprintf(" %s %s", v, argVal)
				}
			}
		case []interface{}:
			// For optional parameters or flags like ["ex", "EX", "px", "PX"] and any order
			for i := 0; i < len(v); i += 2 {
				paramKey := v[i].(string)
				redisKeyword := v[i+1].(string)
				if argVal, ok := args[paramKey]; ok {
					// Add only the Redis keyword if it's a flag
					if argVal == redisKeyword {
						result += fmt.Sprintf(" %s", redisKeyword)
					} else {
						result += fmt.Sprintf(" %s %s", redisKeyword, argVal)
					}
				}
			}
		}
	}

	return result, nil
}

// encodeArgs encodes arguments into a Redis command string based on command metadata.
func encodeArgs(args map[string]interface{}, meta DiceDBAdapterMeta) (string, error) {
	// Initialize command parts with the command name
	commandParts := []string{meta.Command}

	// Append required arguments
	argNames := make([]string, len(meta.RequiredArgs))
	for argName, position := range meta.RequiredArgs {
		argNames[position.BeginIndex] = argName
	}
	for _, argName := range argNames {
		value, exists := args[argName]
		// fmt.Println(argName, value, exists)
		if !exists {
			continue
		}
		switch v := value.(type) {
		case string:
			fmt.Println("value is string", value)
			commandParts = append(commandParts, v) // Add the value
		default:
			fmt.Println("value is possible list", value)
			s, err := flattenStringArgs(value)
			if err != nil {
				return "", err
			}
			commandParts = append(commandParts, s) // Add the values
		}
	}
	// if command can have subcommands
	if len(meta.Subcommands) > 0 {
		// Append subcommands
		if subCommand, exists := args["subcommands"]; exists {
			for _, subcommand := range subCommand.([]map[string]interface{}) {
				subcommandName := subcommand["subcommand"].(string)
				commandParts = append(commandParts, subcommandName)
				delete(subcommand, "subcommand")
				subCommandStrings, err := encodeArgs(subcommand, meta.Subcommands[subcommandName])
				if err != nil {
					return "", err
				}
				commandParts = append(commandParts, subCommandStrings)

			}
		}
	}

	// Append flags
	for flagName, flagKey := range meta.Flags {
		if _, exists := args[flagName]; exists {
			commandParts = append(commandParts, flagKey) // Add the flag
		}
	}

	// Append optional arguments
	for optArgName, optArgKey := range meta.OptionalArgs {
		if value, exists := args[optArgName]; exists {
			commandParts = append(commandParts, optArgKey, value.(string)) // Add key-value pair
		}
	}
	// Join command parts to form the complete command string
	return strings.TrimSpace(strings.Join(commandParts, " ")), nil
}

// setEncoder encodes the SET command.
func setEncoder(args map[string]interface{}) (string, error) {
	return encodeArgs(args, DiceCmdAdapters["SET"])
}

// getEncoder encodes the GET command.
func getEncoder(args map[string]interface{}) (string, error) {
	return encodeArgs(args, DiceCmdAdapters["GET"])
}

// delEncoder encodes the DEL command.
func delEncoder(args map[string]interface{}) (string, error) {
	return encodeArgs(args, DiceCmdAdapters["DEL"])
}

func mgetEncoder(args map[string]interface{}) (string, error) {
	return encodeArgs(args, DiceCmdAdapters["MGET"])
}

func msetEncoder(args map[string]interface{}) (string, error) {
	return encodeArgs(args, DiceCmdAdapters["MSET"])
}

func bitopEncoder(args map[string]interface{}) (string, error) {
	return encodeArgs(args, DiceCmdAdapters["BITOP"])
}

func bitfieldEncoder(args map[string]interface{}) (string, error) {
	return encodeArgs(args, DiceCmdAdapters["BITFIELD"])
}
