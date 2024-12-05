package parsers

func ParseResponse(response interface{}) interface{} {
	switch response := response.(type) {
	case float64:
		return int64(response)
	case nil:
		return "(nil)"
	default:
		return response
	}
}
