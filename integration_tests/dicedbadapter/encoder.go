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

// encodeArgs encodes arguments into a Redis command string based on command metadata.
func encodeArgs(args map[string]interface{}, meta DiceDBAdapterMeta) (string, error) {
	// Initialize command parts with the command name
	commandParts := []string{meta.Command}
	fromEnd := 0
	fromStart := 0
	// Append required arguments
	allArgNames := make([]string, len(meta.RequiredArgs))
	for argName, position := range meta.RequiredArgs {
		if position.BeginIndex < 0 {
			fromEnd++
			continue
		}
		fromStart++
		allArgNames[position.BeginIndex] = argName
	}
	argNames := allArgNames[:fromStart]
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
	// Append flags after sorting them based on their index
	allFlags := make([]string, len(meta.Flags))
	for flagName, flagIndex := range meta.Flags {
		allFlags[flagIndex] = flagName
	}
	for _, flagName := range allFlags {
		if _, exists := args[flagName]; exists {
			commandParts = append(commandParts, flagName) // Add the flag
		}
	}

	allOptionalArgs := make([]string, len(meta.OptionalArgs))
	for optArgName, optArgIndex := range meta.OptionalArgs {
		allOptionalArgs[optArgIndex] = optArgName
	}
	for _, optArgName := range allOptionalArgs {
		if value, exists := args[optArgName]; exists {
			commandParts = append(commandParts, optArgName, value.(string)) // Add key-value pair
		}
	}
	// Append optional arguments
	// for optArgName, _ := range meta.OptionalArgs {
	// 	if value, exists := args[optArgName]; exists {
	// 		commandParts = append(commandParts, optArgName, value.(string)) // Add key-value pair
	// 	}
	// }

	// process the required args that are not yet processed
	// Append required arguments
	argNames = make([]string, fromEnd)
	for argName, position := range meta.RequiredArgs {
		if position.BeginIndex < 0 {
			argNames[fromEnd+position.BeginIndex] = argName
		}
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

func zaddEncoder(args map[string]interface{}) (string, error) {
	return encodeArgs(args, DiceCmdAdapters["ZADD"])
}
