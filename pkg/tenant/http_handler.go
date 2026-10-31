package tenant

import (
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/internal/net"
)

type httpHandler struct {
	svc *service
}

func CreateHTTPHandler(svc *service) *httpHandler {
	return &httpHandler{svc: svc}
}

func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint(id), true
}

func (this *httpHandler) createTenant(c *gin.Context) {
	actor, ok := net.GetActor(c)
	if !ok {
		c.IndentedJSON(401, net.ErrorResponse(common.ErrInvalidCredentials, nil))
		return
	}
	var body createTenantReq
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Println("failed to bind json", "err", err.Error())
		c.IndentedJSON(400, net.ErrorResponse(common.ErrInvalidRequest, nil))
		return
	}
	net.Respond(c, this.svc.createTenant(actor, body), "tenant created")
}

func (this *httpHandler) getTenants(c *gin.Context) {
	actor, ok := net.GetActor(c)
	if !ok {
		c.IndentedJSON(401, net.ErrorResponse(common.ErrInvalidCredentials, nil))
		return
	}
	var query getTenantsReq
	if err := c.ShouldBindQuery(&query); err != nil {
		log.Println("failed to bind query", "err", err.Error())
		c.IndentedJSON(400, net.ErrorResponse(common.ErrInvalidRequest, nil))
		return
	}
	net.Respond(c, this.svc.getTenants(actor, query), "")
}

func (this *httpHandler) getTenantInfo(c *gin.Context) {
	actor, ok := net.GetActor(c)
	if !ok {
		c.IndentedJSON(401, net.ErrorResponse(common.ErrInvalidCredentials, nil))
		return
	}
	id, ok := parseID(c)
	if !ok {
		c.IndentedJSON(400, net.ErrorResponse(common.ErrInvalidRequest, nil))
		return
	}
	net.Respond(c, this.svc.getTenantInfo(actor, id), "")
}

func (this *httpHandler) updateTenantInfo(c *gin.Context) {
	actor, ok := net.GetActor(c)
	if !ok {
		c.IndentedJSON(401, net.ErrorResponse(common.ErrInvalidCredentials, nil))
		return
	}
	id, ok := parseID(c)
	if !ok {
		c.IndentedJSON(400, net.ErrorResponse(common.ErrInvalidRequest, nil))
		return
	}
	var body updateTenantReq
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Println("failed to bind json", "err", err.Error())
		c.IndentedJSON(400, net.ErrorResponse(common.ErrInvalidRequest, nil))
		return
	}
	net.Respond(c, this.svc.updateTenantInfo(actor, id, body), "tenant updated")
}
