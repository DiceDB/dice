// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package utils

// This method returns the path, and a boolean value telling if the path provided follows Legacy Path Syntax or JSONPath syntax.
// JSON knows which syntax to use depending on the first character of the path query.
// If the query starts with the character $, it uses JSONPath syntax. Otherwise, it defaults to the legacy path syntax.
// A JSONPath query can resolve to several locations in a JSON document.
// In this case, the JSON commands apply the operation to every possible location. This is a major improvement over legacy path queries, which only operate on the first path.
func ParseInputJSONPath(path string) (string, bool) {
	isLegacyPath := path[0] != '$'

	// Handle . path error
	if len(path) == 1 && path[0] == '.' {
		path = "$"
	}
	return path, isLegacyPath
}
