package net

func ErrorResponse(message string, data any) map[string]any {
	return map[string]any{
		"status":  "error",
		"message": message,
		"data":    data,
	}
}

func SuccessResponse(data any, message string) map[string]any {
	return map[string]any{
		"status":  "success",
		"message": message,
		"data":    data,
	}
}
