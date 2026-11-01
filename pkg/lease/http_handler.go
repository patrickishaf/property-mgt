package lease

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

// requestActor returns the authenticated actor and the requested numeric path
// ids, writing an error response and returning ok=false if any is invalid.
func requestActor(c *gin.Context, params ...string) (*common.Actor, []uint, bool) {
	actor, ok := net.GetActor(c)
	if !ok {
		c.IndentedJSON(401, net.ErrorResponse(common.ErrInvalidCredentials, nil))
		return nil, nil, false
	}
	ids := make([]uint, 0, len(params))
	for _, name := range params {
		id, err := strconv.ParseUint(c.Param(name), 10, 64)
		if err != nil || id == 0 {
			c.IndentedJSON(400, net.ErrorResponse(common.ErrInvalidRequest, nil))
			return nil, nil, false
		}
		ids = append(ids, uint(id))
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

func (this *httpHandler) getLeases(c *gin.Context) {
	actor, _, ok := requestActor(c)
	var query getLeasesReq
	if !ok || !bindQuery(c, &query) {
		return
	}
	net.Respond(c, this.svc.getLeases(actor, query), "")
}

func (this *httpHandler) getLeaseById(c *gin.Context) {
	actor, ids, ok := requestActor(c, "id")
	if !ok {
		return
	}
	net.Respond(c, this.svc.getLeaseById(actor, ids[0]), "")
}

func (this *httpHandler) getLeaseInstallments(c *gin.Context) {
	actor, ids, ok := requestActor(c, "id")
	if !ok {
		return
	}
	net.Respond(c, this.svc.getLeaseInstallments(actor, ids[0]), "")
}

func (this *httpHandler) createLeaseInstallment(c *gin.Context) {
	actor, ids, ok := requestActor(c, "id")
	var body setInstallmentsReq
	if !ok || !bindJSON(c, &body) {
		return
	}
	net.Respond(c, this.svc.createLeaseInstallment(actor, ids[0], body), "installments set")
}

func (this *httpHandler) createLease(c *gin.Context) {
	actor, _, ok := requestActor(c)
	var body createLeaseReq
	if !ok || !bindJSON(c, &body) {
		return
	}
	net.Respond(c, this.svc.createLease(actor, body), "lease created")
}

func (this *httpHandler) updateLease(c *gin.Context) {
	actor, ids, ok := requestActor(c, "id")
	var body updateLeaseReq
	if !ok || !bindJSON(c, &body) {
		return
	}
	net.Respond(c, this.svc.updateLease(actor, ids[0], body), "lease updated")
}

func (this *httpHandler) getInstallments(c *gin.Context) {
	actor, _, ok := requestActor(c)
	var query getInstallmentsReq
	if !ok || !bindQuery(c, &query) {
		return
	}
	net.Respond(c, this.svc.getInstallments(actor, query), "")
}

func (this *httpHandler) updateInstallment(c *gin.Context) {
	actor, ids, ok := requestActor(c, "id")
	var body updateInstallmentReq
	if !ok || !bindJSON(c, &body) {
		return
	}
	net.Respond(c, this.svc.updateInstallment(actor, ids[0], body), "installment updated")
}

func (this *httpHandler) getDue(c *gin.Context) {
	actor, _, ok := requestActor(c)
	var query getDueReq
	if !ok || !bindQuery(c, &query) {
		return
	}
	net.Respond(c, this.svc.getDue(actor, query), "")
}

func (this *httpHandler) runExpiry(c *gin.Context) {
	net.Respond(c, this.svc.runExpiry(), "expiry job ran")
}
