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
					if strings.ToLower(v) != True {
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
					if strings.ToLower(value) != True {
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
