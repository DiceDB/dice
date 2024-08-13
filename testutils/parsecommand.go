package testutils

import (
	"strings"

	"github.com/dicedb/dice/internal/constants"
)

func ParseCommand(cmd string) []string {
	var args []string
	var current string
	var inQuotes bool

	for _, char := range cmd {
		switch char {
		case ' ':
			if inQuotes {
				current += string(char)
			} else if current != constants.EmptyStr {
				args = append(args, current)
				current = constants.EmptyStr
			}
		case '"':
			inQuotes = !inQuotes
			current += string(char)
		default:
			current += string(char)
		}
	}

	if current != constants.EmptyStr {
		args = append(args, current)
	}

	// Remove quotes from each argument
	for i, arg := range args {
		args[i] = strings.Trim(arg, `"`)
	}

	return args
}
