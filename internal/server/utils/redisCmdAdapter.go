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

func ParseHTTPRequest(r *http.Request) (*cmd.RedisCmd, error) {
	prefix := strings.TrimPrefix(r.URL.Path, "/")
	if prefix == "" {
		return nil, errors.New("invalid command")
	}

	prefix = strings.ToUpper(prefix)
	var args []string

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
			priorityKeys := []string{"key", "field", "path", "value"}
			for _, key := range priorityKeys {
				if val, exists := jsonBody[key]; exists {
					args = append(args, fmt.Sprintf("%v", val))
					delete(jsonBody, key)
				}
			}

			// Process remaining keys
			for k, v := range jsonBody {
				// Handle unary operations like 'nx' where value is "true"
				if valStr, ok := v.(string); ok && valStr == "true" {
					args = append(args, k)
				} else {
					switch val := v.(type) {
					case map[string]interface{}, []interface{}:
						jsonValue, err := json.Marshal(val)
						if err != nil {
							return nil, err
						}
						args = append(args, string(jsonValue))
					default:
						args = append(args, fmt.Sprintf("%v", val))
					}
				}
			}
		}
	}

	// Step 2: Return the constructed Redis command
	return &cmd.RedisCmd{
		Cmd:  prefix,
		Args: args,
	}, nil
}
