package utils

type ResponseResult struct {
	Message string         `json:"message"`
	Data    map[string]any `json:"data"`
}

func Response(message string, data map[string]any) *ResponseResult {
	return &ResponseResult{
		Message: message,
		Data:    data,
	}
}
