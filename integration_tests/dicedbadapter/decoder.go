package dicedbadapter

import (
	"fmt"
	"strings"
)

// decodeArgs decodes a Redis command string based on the command metadata.
func decodeArgs(commandParts []string, meta DiceDBAdapterMeta) (map[string]interface{}, error) {
	// Ensure command name matches and that we have sufficient parts
	if len(commandParts) == 0 {
		return nil, fmt.Errorf("empty command parts")
	}

	decodedArgs := make(map[string]interface{})
	argIndex := 0 // Start after the command name

	// Decode required arguments with positive begin index
	fromStart := 0
	fromEnd := 0
	allRequiredArgs := make([]string, len(meta.RequiredArgs))
	for argName, position := range meta.RequiredArgs {
		if position.BeginIndex >= 0 {
			fromStart++
			allRequiredArgs[position.BeginIndex] = argName
		} else {
			fromEnd++
		}
	}
	allFrontRequiredArgs := allRequiredArgs[:fromStart]
	// fmt.Println(allFrontRequiredArgs)
	for _, argName := range allFrontRequiredArgs {
		position := meta.RequiredArgs[argName]
		// the required argument is a list till the end of the command
		if position.EndIndex == -1 {
			position.EndIndex = len(commandParts) - 1
			argVal := make([]interface{}, 0)
			if position.Step > 1 {
				for i := position.BeginIndex; i <= position.EndIndex; i += position.Step {
					temp := make([]interface{}, 0, position.Step)
					for j := 0; ; j++ {
						// if we have reached the end of the command
						// and we still need more arguments
						// then we append an empty string
						if i+j >= len(commandParts) && j < position.Step {
							temp = append(temp, "")
							break
						}
						// if we have reached the end of the command
						if j >= position.Step {
							break
						}
						temp = append(temp, commandParts[i+j])
					}
					argVal = append(argVal, temp)
				}
				decodedArgs[argName] = argVal
			} else {
				for i := position.BeginIndex; i <= position.EndIndex; i += position.Step {
					argVal = append(argVal, commandParts[i])
				}
				decodedArgs[argName] = argVal
			}
			argIndex += position.EndIndex - position.BeginIndex + 1
			// the required argument is a string
		} else if position.BeginIndex < len(commandParts) {
			decodedArgs[argName] = commandParts[position.BeginIndex]
			argIndex++
		} else {
			// missing required argument
			// we do not validate the command here
			// only server side should validate the command
			decodedArgs[argName] = ""
			argIndex++
		}
	}
	// fmt.Println(decodedArgs, argIndex, commandParts)
	// Decode Subcommands
	if len(meta.Subcommands) > 0 {
		decodedArgs["subcommands"] = make([]map[string]interface{}, 0)
		for argIndex < len(commandParts) {
			subcommand := strings.ToLower(commandParts[argIndex])
			argIndex++
			if _, exists := meta.Subcommands[subcommand]; !exists {
				decodedArgs[subcommand] = ""
				continue
			}
			subcommandMeta := meta.Subcommands[subcommand]
			subCommandParts := make([]string, 0)
			for j := 0; j < len(subcommandMeta.RequiredArgs); j++ {
				if argIndex >= len(commandParts) {
					subCommandParts = append(subCommandParts, "")
					continue
				}
				subCommandParts = append(subCommandParts, commandParts[argIndex])
				argIndex++
			}
			subCommandArgs, err := decodeArgs(subCommandParts, subcommandMeta)
			if err != nil {
				return nil, err
			}
			subCommandArgs["subcommand"] = subcommand
			decodedArgs["subcommands"] = append(decodedArgs["subcommands"].([]map[string]interface{}), subCommandArgs)
		}
	}
	// fmt.Println(decodedArgs, argIndex, commandParts)

	for ; argIndex < len(commandParts); argIndex++ {
		// Decode optional arguments
		arg := strings.ToLower(commandParts[argIndex])
		if _, exists := meta.OptionalArgs[arg]; exists {
			if argIndex+1 < len(commandParts) {
				decodedArgs[arg] = commandParts[argIndex+1]
			} else {
				decodedArgs[arg] = ""
			}
			argIndex++
		} else if _, exists := meta.Flags[arg]; exists {
			decodedArgs[arg] = arg
		} else {
			break
		}
	}
	// fmt.Println(decodedArgs, argIndex, commandParts)
	allEndRequiredArgs := make([]string, fromEnd)
	for argName, position := range meta.RequiredArgs {
		if position.BeginIndex < 0 {
			allEndRequiredArgs[fromEnd+position.BeginIndex] = argName
		}
	}
	// fmt.Println(allEndRequiredArgs)

	for _, argName := range allEndRequiredArgs {
		position := meta.RequiredArgs[argName]
		// the required argument is a list till the end of the command
		if position.EndIndex == -1 {
			position.EndIndex = len(commandParts) - 1
			argVal := make([]interface{}, 0)
			if position.Step > 1 {
				for ; argIndex <= position.EndIndex; argIndex += position.Step {
					temp := make([]interface{}, 0, position.Step)
					for j := 0; ; j++ {
						// if we have reached the end of the command
						// and we still need more arguments
						// then we append an empty string
						if argIndex+j >= len(commandParts) && j < position.Step {
							temp = append(temp, "")
							break
						}
						// if we have reached the end of the command
						if j >= position.Step {
							break
						}
						temp = append(temp, commandParts[argIndex+j])
					}
					argVal = append(argVal, temp)
				}
				decodedArgs[argName] = argVal
			} else {
				for ; argIndex <= position.EndIndex; argIndex += position.Step {
					argVal = append(argVal, commandParts[argIndex])
				}
				decodedArgs[argName] = argVal
			}
			argIndex += position.EndIndex - position.BeginIndex + 1
		} else if position.BeginIndex < len(commandParts) {
			decodedArgs[argName] = commandParts[position.BeginIndex]
			argIndex++
		} else {
			// missing required argument
			// we do not validate the command here
			// only server side should validate the command
			decodedArgs[argName] = ""
		}
	}
	return decodedArgs, nil
}

// setDecoder decodes the SET command.
func setDecoder(commandParts []string) (map[string]interface{}, error) {
	return decodeArgs(commandParts, DiceCmdAdapters["SET"])
}

// getDecoder decodes the GET command.
func getDecoder(parts []string) (map[string]interface{}, error) {
	return decodeArgs(parts, DiceCmdAdapters["GET"])
}

// delDecoder decodes the DEL command.
func delDecoder(parts []string) (map[string]interface{}, error) {
	return decodeArgs(parts, DiceCmdAdapters["DEL"])
}

func mgetDecoder(commandParts []string) (map[string]interface{}, error) {
	return decodeArgs(commandParts, DiceCmdAdapters["MGET"])
}

func msetDecoder(commandParts []string) (map[string]interface{}, error) {
	return decodeArgs(commandParts, DiceCmdAdapters["MSET"])
}

func bitopDecoder(commandParts []string) (map[string]interface{}, error) {
	return decodeArgs(commandParts, DiceCmdAdapters["BITOP"])
}

func bitfieldDecoder(commandParts []string) (map[string]interface{}, error) {
	return decodeArgs(commandParts, DiceCmdAdapters["BITFIELD"])
}

func zaddDecoder(commandParts []string) (map[string]interface{}, error) {
	return decodeArgs(commandParts, DiceCmdAdapters["ZADD"])
}
