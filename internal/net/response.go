package net

import (
	"github.com/gin-gonic/gin"
	"github.com/patrickishaf/property-mgt/common"
)

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

// Respond writes a ServiceResponse using the standard response envelope.
func Respond(c *gin.Context, res common.ServiceResponse, message string) {
	if res.Err != "" {
		c.IndentedJSON(res.StatusCode, ErrorResponse(res.Err, nil))
	} else {
		c.IndentedJSON(res.StatusCode, SuccessResponse(res.Data, message))
	}
}
