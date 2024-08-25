package testutils

import (
	"strings"
	"unicode/utf8"
)

func ParseCommand(cmd string) []string {
	var args []string
	var builder strings.Builder
	var inQuotes bool

	flushBuilder := func() {
		if builder.Len() > 0 {
			args = append(args, builder.String())
			builder.Reset()
		}
	}

	for cmd != "" {
		r, size := utf8.DecodeRuneInString(cmd)
		switch {
		case r == utf8.RuneError && size == 1:
			// Invalid UTF-8 sequence, treat each byte as a separate character
			builder.WriteByte(cmd[0])
			cmd = cmd[1:]
		case r == ' ' && !inQuotes:
			flushBuilder()
			cmd = cmd[size:]
		case r == '"':
			inQuotes = !inQuotes
			builder.WriteRune(r)
			cmd = cmd[size:]
		default:
			builder.WriteString(cmd[:size])
			cmd = cmd[size:]
		}
	}

	flushBuilder()

	// Remove quotes from each argument
	for i, arg := range args {
		args[i] = trimQuotes(arg)
	}

	return args
}

func trimQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}
