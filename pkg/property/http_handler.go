package property

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
	return &httpHandler{
		svc: svc,
	}
}

func parseUintParam(c *gin.Context, name string) (uint, bool) {
	id, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint(id), true
}

// requestActor returns the authenticated actor and the requested path ids, writing
// an error response and returning ok=false if any is missing or invalid.
func requestActor(c *gin.Context, params ...string) (*common.Actor, []uint, bool) {
	actor, ok := net.GetActor(c)
	if !ok {
		c.IndentedJSON(401, net.ErrorResponse(common.ErrInvalidCredentials, nil))
		return nil, nil, false
	}
	ids := make([]uint, 0, len(params))
	for _, name := range params {
		id, ok := parseUintParam(c, name)
		if !ok {
			c.IndentedJSON(400, net.ErrorResponse(common.ErrInvalidRequest, nil))
			return nil, nil, false
		}
		ids = append(ids, id)
	}
	return actor, ids, true
}

func bindJSON(c *gin.Context, body any) bool {
	if err := c.ShouldBindJSON(body); err != nil {
		log.Println("failed to bind json", "err", err.Error())
		c.IndentedJSON(400, net.ErrorResponse(common.ErrInvalidRequest, nil))
		return false
	}
	return true
}

func bindQuery(c *gin.Context, query any) bool {
	if err := c.ShouldBindQuery(query); err != nil {
		log.Println("failed to bind query", "err", err.Error())
		c.IndentedJSON(400, net.ErrorResponse(common.ErrInvalidRequest, nil))
		return false
	}
	return true
}

func (this *httpHandler) createProperty(c *gin.Context) {
	actor, _, ok := requestActor(c)
	var body createPropertyReq
	if !ok || !bindJSON(c, &body) {
		return
	}
	net.Respond(c, this.svc.createProperty(actor, body), "property created")
}

func (this *httpHandler) getProperties(c *gin.Context) {
	actor, _, ok := requestActor(c)
	var query getPropertiesReq
	if !ok || !bindQuery(c, &query) {
		return
	}
	net.Respond(c, this.svc.getProperties(actor, query), "")
}

func (this *httpHandler) getPropertyInfo(c *gin.Context) {
	actor, ids, ok := requestActor(c, "id")
	if !ok {
		return
	}
	net.Respond(c, this.svc.getPropertyInfo(actor, ids[0]), "")
}

func (this *httpHandler) updatePropertyInfo(c *gin.Context) {
	actor, ids, ok := requestActor(c, "id")
	var body updatePropertyReq
	if !ok || !bindJSON(c, &body) {
		return
	}
	net.Respond(c, this.svc.updatePropertyInfo(actor, ids[0], body), "property updated")
}

func (this *httpHandler) archiveProperty(c *gin.Context) {
	actor, ids, ok := requestActor(c, "id")
	if !ok {
		return
	}
	net.Respond(c, this.svc.archiveProperty(actor, ids[0]), "property archived")
}

func (this *httpHandler) assignPropertyManager(c *gin.Context) {
	actor, ids, ok := requestActor(c, "id")
	var body assignManagerReq
	if !ok || !bindJSON(c, &body) {
		return
	}
	net.Respond(c, this.svc.assignPropertyManager(actor, ids[0], body), "manager assigned")
}

func (this *httpHandler) unassignPropertyManager(c *gin.Context) {
	actor, ids, ok := requestActor(c, "id")
	if !ok {
		return
	}
	net.Respond(c, this.svc.unassignPropertyManager(actor, ids[0]), "manager unassigned")
}

func (this *httpHandler) getUnits(c *gin.Context) {
	actor, ids, ok := requestActor(c, "id")
	if !ok {
		return
	}
	net.Respond(c, this.svc.getUnits(actor, ids[0]), "")
}

func (this *httpHandler) createUnit(c *gin.Context) {
	actor, ids, ok := requestActor(c, "id")
	var body createUnitReq
	if !ok || !bindJSON(c, &body) {
		return
	}
	net.Respond(c, this.svc.createUnit(actor, ids[0], body), "unit created")
}

func (this *httpHandler) updateUnit(c *gin.Context) {
	actor, ids, ok := requestActor(c, "id", "unitId")
	var body updateUnitReq
	if !ok || !bindJSON(c, &body) {
		return
	}
	net.Respond(c, this.svc.updateUnit(actor, ids[0], ids[1], body), "unit updated")
}

func (this *httpHandler) deleteUnit(c *gin.Context) {
	actor, ids, ok := requestActor(c, "id", "unitId")
	if !ok {
		return
	}
	net.Respond(c, this.svc.deleteUnit(actor, ids[0], ids[1]), "unit deleted")
}

func (this *httpHandler) createPropertyManager(c *gin.Context) {
	actor, _, ok := requestActor(c)
	var body createManagerReq
	if !ok || !bindJSON(c, &body) {
		return
	}
	net.Respond(c, this.svc.createPropertyManager(actor, body), "manager created")
}

func (this *httpHandler) getPropertyManagers(c *gin.Context) {
	actor, _, ok := requestActor(c)
	var query pageReq
	if !ok || !bindQuery(c, &query) {
		return
	}
	net.Respond(c, this.svc.getPropertyManagers(actor, query), "")
}

func (this *httpHandler) getPropertyManager(c *gin.Context) {
	actor, ids, ok := requestActor(c, "id")
	if !ok {
		return
	}
	net.Respond(c, this.svc.getPropertyManager(actor, ids[0]), "")
}

func (this *httpHandler) updatePropertyManager(c *gin.Context) {
	actor, ids, ok := requestActor(c, "id")
	var body updateManagerReq
	if !ok || !bindJSON(c, &body) {
		return
	}
	net.Respond(c, this.svc.updatePropertyManager(actor, ids[0], body), "manager updated")
}

func (this *httpHandler) removePropertyManager(c *gin.Context) {
	actor, ids, ok := requestActor(c, "id")
	if !ok {
		return
	}
	net.Respond(c, this.svc.removePropertyManager(actor, ids[0]), "manager removed")
}
