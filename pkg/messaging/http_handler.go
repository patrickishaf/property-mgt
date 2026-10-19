package messaging

import "github.com/gin-gonic/gin"

type httpHandler struct {
	svc *service
}

func CreateHTTPHandler(svc *service) *httpHandler {
	return &httpHandler{
		svc: svc,
	}
}

func (this *httpHandler) getMessageRecipients(c *gin.Context) {}

func (this *httpHandler) getMessageInfo(c *gin.Context) {}

func (this *httpHandler) getMessages(c *gin.Context) {}

func (this *httpHandler) getTemplates(c *gin.Context) {}

func (this *httpHandler) sendMessage(c *gin.Context) {}
