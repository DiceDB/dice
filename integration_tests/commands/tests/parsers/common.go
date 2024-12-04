package parsers

func ParseResponse(response interface{}) interface{} {
	// convert the output to the int64 if it is float64
	switch response.(type) {
	case float64:
		return int64(response.(float64))
	case nil:
		return "(nil)"
	default:
		return response
	}
}
