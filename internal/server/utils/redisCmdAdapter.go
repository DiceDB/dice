package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dicedb/dice/internal/cmd"
)

const (
	Command   = "command"
	Key       = "key"
	KeyPrefix = "key_prefix"
	Field     = "field"
	Path      = "path"
	Value     = "value"
	KeyValues = "key_values"
	True      = "true"
)

func ParseHTTPRequest(r *http.Request) (*cmd.RedisCmd, error) {
	command := strings.TrimPrefix(r.URL.Path, "/")
	if command == "" {
		return nil, errors.New("invalid command")
	}

	command = strings.ToUpper(command)
	var args []string

	// Extract query parameters
	queryParams := r.URL.Query()
	keyPrefix := queryParams.Get(KeyPrefix)

	if keyPrefix != "" && command == "JSON.INGEST" {
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

			// Define keys to exclude and process their values first
			// Update as we support more commands
			var priorityKeys = [5]string{
				Key,
				Field,
				Path,
				Value,
				KeyValues,
			}
			for _, key := range priorityKeys {
				if val, exists := jsonBody[key]; exists {
					args = append(args, fmt.Sprintf("%v", val))
					delete(jsonBody, key)
				}
			}

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

	// Step 2: Return the constructed Redis command
	return &cmd.RedisCmd{
		Cmd:  command,
		Args: args,
	}, nil
}

func ParseWebsocketMessage(msg []byte) (*cmd.RedisCmd, error) {
	var command string
	var args []string

	// parse msg to json
	var jsonBody map[string]interface{}
	if err := json.Unmarshal(msg, &jsonBody); err != nil {
		return nil, fmt.Errorf("error parsing message: %v", err)
	}

	// extract command
	if comStr, exists := jsonBody[Command]; exists {
		var ok bool
		command, ok = comStr.(string)
		if !ok {
			return nil, fmt.Errorf("error typecasting command to string")
		}
		command = strings.ToUpper(command)
	} else {
		return nil, fmt.Errorf("error extracting command")
	}
	delete(jsonBody, Command)

	// extract priority keys
	var priorityKeys = [6]string{
		KeyPrefix,
		Key,
		Field,
		Path,
		Value,
		KeyValues,
	}
	for _, key := range priorityKeys {
		if val, exists := jsonBody[key]; exists {
			args = append(args, fmt.Sprintf("%v", val))
			delete(jsonBody, key)
		} else if command == "JSON.INGEST" && key == KeyPrefix {
			// add empty key prefix
			args = append(args, "")
		}
	}

	// process remaining keys
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

	return &cmd.RedisCmd{
		Cmd:  command,
		Args: args,
	}, nil
}
