package settings

import "github.com/gin-gonic/gin"

type httpHandler struct {
	svc *service
}

func CreateHTTPHandler(svc *service) *httpHandler {
	return &httpHandler{svc: svc}
}

func (this *httpHandler) getUserSettings(c *gin.Context) {}

func (this *httpHandler) updateProfileSettings(c *gin.Context) {}

func (this *httpHandler) updatePreferences(c *gin.Context) {}
