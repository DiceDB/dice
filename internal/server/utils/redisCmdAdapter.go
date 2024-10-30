package utils

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/dicedb/dice/internal/cmd"
	diceerrors "github.com/dicedb/dice/internal/errors"
)

const (
	Key              = "key"
	Keys             = "keys"
	KeyPrefix        = "key_prefix"
	Field            = "field"
	Path             = "path"
	Value            = "value"
	Values           = "values"
	User             = "user"
	Password         = "password"
	Seconds          = "seconds"
	KeyValues        = "key_values"
	True             = "true"
	QwatchQuery      = "query"
	Offset           = "offset"
	Member           = "member"
	Members          = "members"
	Index            = "index"
	JSON             = "json"
	QWatch           = "Q.WATCH"
	ABORT            = "ABORT"
	IsByteEncodedVal = "isByteEncodedVal"
)

func ParseHTTPRequest(r *http.Request) (*cmd.DiceDBCmd, error) {
	commandParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if len(commandParts) == 0 {
		return nil, errors.New("invalid command")
	}

	command := strings.ToUpper(commandParts[0])

	var subcommand string
	if len(commandParts) > 1 {
		subcommand = strings.ToUpper(commandParts[1])
	}

	var args []string

	// Handle subcommand and multiple arguments
	if subcommand != "" {
		args = append(args, subcommand)
	}

	// Extract query parameters
	queryParams := r.URL.Query()
	keyPrefix := queryParams.Get(KeyPrefix)

	if keyPrefix != "" && command == JSONIngest {
		args = append(args, keyPrefix)
	}
	// Step 1: Handle JSON body if present
	if r.Body != nil {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}

		if len(body) > 0 {
			var jsonBody map[string]interface{}
			if err := json.Unmarshal(body, &jsonBody); err != nil {
				return nil, err
			}

			if len(jsonBody) == 0 && command != ABORT {
				return nil, fmt.Errorf("empty JSON object")
			}

			// Define keys to exclude and process their values first
			// Update as we support more commands
			processPriorityKeys(jsonBody, &args)

			// Process remaining keys in the JSON body
			for key, val := range jsonBody {
				switch v := val.(type) {
				case string:
					// Handle unary operations like 'nx' where value is "true"
					args = append(args, key)
					if !strings.EqualFold(v, True) {
						args = append(args, v)
					}
				case map[string]interface{}, []interface{}:
					// Marshal nested JSON structures back into a string
					jsonValue, err := json.Marshal(v)
					if err != nil {
						return nil, err
					}
					args = append(args, string(jsonValue))
				default:
					args = append(args, key)
					// Append other types as strings
					value := fmt.Sprintf("%v", v)
					if !strings.EqualFold(value, True) {
						args = append(args, value)
					}
				}
			}
		}
	}

	// Step 2: Return the constructed DiceDB command
	return &cmd.DiceDBCmd{
		Cmd:  command,
		Args: args,
	}, nil
}

func ParseWebsocketMessage(msg []byte) (*cmd.DiceDBCmd, error) {
	cmdStr := strings.TrimSpace(string(msg))
	if cmdStr == "" {
		return nil, diceerrors.ErrEmptyCommand
	}

	var command string
	idx := strings.Index(cmdStr, " ")
	// handle commands with no args
	if idx == -1 {
		command = strings.ToUpper(cmdStr)
		return &cmd.DiceDBCmd{
			Cmd:  command,
			Args: nil,
		}, nil
	}

	// handle commands with args
	command = strings.ToUpper(cmdStr[:idx])
	cmdStr = cmdStr[idx+1:]

	regexPattern := `"(.*?)"|'(.*?)'|(\S+)`
	re := regexp.MustCompile(regexPattern)
	matches := re.FindAllStringSubmatch(cmdStr, -1)

	var cmdArr []string // args

	// handle qwatch commands
	if command == QWatch {
		// remove quotes from query string
		cmdStr, err := strconv.Unquote(cmdStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing q.watch query: %v", err)
		}
		cmdArr = []string{cmdStr}
	} else {
		// handle other commands
		for _, match := range matches {
			if match[1] != "" {
				cmdArr = append(cmdArr, match[1])
			} else if match[2] != "" {
				cmdArr = append(cmdArr, match[2])
			} else {
				cmdArr = append(cmdArr, match[3])
			}
		}
	}

	// if key prefix is empty for JSON.INGEST command
	// add "" to cmdArr
	if command == JSONIngest && len(cmdArr) == 2 {
		cmdArr = append([]string{""}, cmdArr...)
	}

	return &cmd.DiceDBCmd{
		Cmd:  command,
		Args: cmdArr,
	}, nil
}

func processPriorityKeys(jsonBody map[string]interface{}, args *[]string) {
	for _, key := range getPriorityKeys() {
		if val, exists := jsonBody[key]; exists {
			switch key {
			case Keys, Members:
				for _, v := range val.([]interface{}) {
					*args = append(*args, fmt.Sprintf("%v", v))
				}
			case JSON:
				jsonValue, _ := json.Marshal(val)
				*args = append(*args, string(jsonValue))
			case KeyValues:
				for k, v := range val.(map[string]interface{}) {
					*args = append(*args, k, fmt.Sprintf("%v", v))
				}
			case Value:
				if _, ok := jsonBody[IsByteEncodedVal]; ok {
					*args = append(*args, formatValue(true, val))
				} else {
					*args = append(*args, formatValue(false, val))
				}
				// Delete the byte encoded val key as it's not required once value is decoded
				delete(jsonBody, IsByteEncodedVal)
			case Values:
				for _, v := range val.([]interface{}) {
					*args = append(*args, fmt.Sprintf("%v", v))
				}
			default:
				*args = append(*args, fmt.Sprintf("%v", val))
			}
			delete(jsonBody, key)
		}
	}
}

func getPriorityKeys() []string {
	return []string{
		Key, Keys, Field, Path, JSON, Index, Value, Values, Seconds, User, Password,
		KeyValues, QwatchQuery, Offset, Member, Members,
	}
}

func formatValue(isByteEncodedVal bool, val interface{}) string {
	switch v := val.(type) {
	case string:
		if isByteEncodedVal && isBase64Encoded(v) {
			decoded, err := base64.StdEncoding.DecodeString(v)
			if err == nil {
				// Replace the base64 string with the decoded `[]byte`
				return string(decoded)
			}
		}
		return v
	default:
		jsonBytes, _ := json.Marshal(v)
		return string(jsonBytes)
	}
}

func isBase64Encoded(s string) bool {
	if len(s)%4 == 0 && s != "" {
		for _, r := range s {
			if !(r >= 'A' && r <= 'Z') && !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') && r != '+' && r != '/' && r != '=' {
				return false
			}
		}
		return true
	}
	return false
}
