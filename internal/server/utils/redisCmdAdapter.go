package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dicedb/dice/internal/cmd"
	"io"
	"net/http"
	"strings"
)

func ParseHTTPRequest(r *http.Request) (*cmd.RedisCmd, error) {
	prefix := strings.TrimPrefix(r.URL.Path, "/")
	if prefix == "" {
		return nil, errors.New("invalid command")
	}

	prefix = strings.ToUpper(prefix)
	var args []string

	// Step 1: Check if query parameters are present
	query := r.URL.Query()
	if len(query) > 0 {
		// Parse from query parameters
		for key, values := range query {
			if len(values) == 1 && values[0] == "" {
				// Treat key as a flag (e.g., "nx")
				args = append(args, key)
			} else {
				for _, value := range values {
					args = append(args, value) // Add values directly
				}
			}
		}
	} else if r.Body != nil {
		// Step 2: Parse from JSON body if present
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}

		if len(body) > 0 {
			var jsonBody map[string]interface{}
			if err := json.Unmarshal(body, &jsonBody); err != nil {
				return nil, err
			}

			// Convert JSON body to arguments
			// Convert JSON body to arguments in the natural order
			for k, v := range jsonBody {
				switch val := v.(type) {
				case map[string]interface{}, []interface{}:
					jsonValue, err := json.Marshal(val)
					if err != nil {
						return nil, err
					}
					args = append(args, k, string(jsonValue))
				default:
					args = append(args, k, fmt.Sprintf("%v", val))
				}
			}
		}
	}

	// Step 3: Return the constructed Redis command
	return &cmd.RedisCmd{
		Cmd:  prefix,
		Args: args,
	}, nil
}
