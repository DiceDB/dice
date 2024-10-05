package utils

func ParseInputJsonPath(path string) (string, bool) {
	isDefinitePath := path[0] == '.'
	if isDefinitePath {
		if len(path) == 1 {
			path = "$"
		} else {
			path = "$" + path
		}
	}
	return path, isDefinitePath
}
