package dicedbadapter

import (
	"fmt"
	"strings"
)

// DiceDBCommand represents a DiceDB command with its route and arguments.
type DiceDBCommand struct {
	Route string                 `json:"route"`
	Body  map[string]interface{} `json:"body"`
}

// CommandEncoder is a function type for encoding a command from map format to string.
type CommandEncoder func(args map[string]interface{}) (string, error)

// CommandDecoder is a function type for decoding a command from string to map format.
type CommandDecoder func(parts []string) (map[string]interface{}, error)

// EncodeCommand encodes a DiceDBCommand to a string using the appropriate encoder.
func EncodeCommand(cmd DiceDBCommand) (string, error) {
	commandKey := strings.Split(cmd.Route, "/")
	if len(commandKey) < 2 {
		return "", fmt.Errorf("command string is empty")
	}
	commandStr := strings.ToUpper(commandKey[1])
	commandData, exists := DiceCmdAdapters[commandStr]
	if !exists {
		return "", fmt.Errorf("command '%s' not supported", cmd.Route)
	}
	return commandData.Encoder(cmd.Body)
}

// DecodeCommand decodes a command string into a DiceDBCommand struct.
func DecodeCommand(commandStr string) (DiceDBCommand, error) {
	parts := strings.Fields(commandStr)
	if len(parts) == 0 {
		return DiceDBCommand{}, fmt.Errorf("command string is empty")
	}

	commandData, exists := DiceCmdAdapters[strings.ToUpper(parts[0])]
	if !exists {
		return DiceDBCommand{}, fmt.Errorf("command '%s' not supported", parts[0])
	}

	body, err := commandData.Decoder(parts[1:])
	if err != nil {
		return DiceDBCommand{}, err
	}

	return DiceDBCommand{
		Route: commandData.Route,
		Body:  body,
	}, nil
}
