package tenant

import "github.com/gin-gonic/gin"

type httpHandler struct {
	svc *service
}

func CreateHTTPHandler(svc *service) *httpHandler {
	return &httpHandler{svc: svc}
}

func (this *httpHandler) createTenant(c *gin.Context) {}

func (this *httpHandler) getTenants(c *gin.Context) {}

func (this *httpHandler) getTenantInfo(c *gin.Context) {}

func (this *httpHandler) updateTenantInfo(c *gin.Context) {}
