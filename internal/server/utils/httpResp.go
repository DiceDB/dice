package utils

const (
	HTTPStatusSuccess string = "success"
	HTTPStatusError   string = "error"
)

type HTTPResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}
