package property

import "github.com/gin-gonic/gin"

type httpHandler struct {
	svc *service
}

func CreateHTTPHandler(svc *service) *httpHandler {
	return &httpHandler{
		svc: svc,
	}
}

func (this *httpHandler) createPropertyManager(c *gin.Context) {}

func (this *httpHandler) getProperties(c *gin.Context) {}

func (this *httpHandler) getPropertyInfo(c *gin.Context) {}

func (this *httpHandler) getPropertyManager(c *gin.Context) {}

func (this *httpHandler) getPropertyManagers(c *gin.Context) {}

func (this *httpHandler) removePropertyManager(c *gin.Context) {}

func (this *httpHandler) updatePropertyInfo(c *gin.Context) {}

func (this *httpHandler) updatePropertyManager(c *gin.Context) {}
