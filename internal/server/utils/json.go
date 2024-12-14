// This file is part of DiceDB.
// Copyright (C) 2024 DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

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
