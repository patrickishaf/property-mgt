package payment

import (
	"encoding/json"
	"io"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/internal/net"
	"github.com/patrickishaf/property-mgt/providers/bachs"
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

func (this *httpHandler) createPayment(c *gin.Context) {
	actor, _, ok := requestActor(c)
	var body createPaymentReq
	if !ok || !bindJSON(c, &body) {
		return
	}
	net.Respond(c, this.svc.createPayment(actor, body), "payment recorded")
}

func (this *httpHandler) getPaymentInfo(c *gin.Context) {
	actor, ids, ok := requestActor(c, "id")
	if !ok {
		return
	}
	net.Respond(c, this.svc.getPaymentInfo(actor, ids[0]), "")
}

func (this *httpHandler) getPayments(c *gin.Context) {
	actor, _, ok := requestActor(c)
	var query getPaymentsReq
	if !ok || !bindQuery(c, &query) {
		return
	}
	net.Respond(c, this.svc.getPayments(actor, query), "")
}

func (this *httpHandler) getTotalPaymentStats(c *gin.Context) {
	actor, _, ok := requestActor(c)
	var query getStatsReq
	if !ok || !bindQuery(c, &query) {
		return
	}
	net.Respond(c, this.svc.getTotalPaymentStats(actor, query), "")
}

func (this *httpHandler) initiateCheckout(c *gin.Context) {
	actor, _, ok := requestActor(c)
	var body initiateCheckoutReq
	if !ok || !bindJSON(c, &body) {
		return
	}
	net.Respond(c, this.svc.initiateCheckout(actor, body), "checkout created")
}

// bachsWebhook receives Bachs checkout events. It is unauthenticated, so every
// delivery must be verified as coming from Bachs before it is trusted.
func (this *httpHandler) bachsWebhook(c *gin.Context) {
	raw, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
	if err != nil {
		c.IndentedJSON(400, net.ErrorResponse(common.ErrInvalidRequest, nil))
		return
	}
	if !this.svc.verifyBachsWebhook(c.Request.Header, raw) {
		c.IndentedJSON(401, net.ErrorResponse(common.ErrInvalidCredentials, nil))
		return
	}
	var event bachs.WebhookEvent
	if err := json.Unmarshal(raw, &event); err != nil || event.Type == "" || event.Data.Reference == "" {
		c.IndentedJSON(400, net.ErrorResponse(common.ErrInvalidRequest, nil))
		return
	}
	net.Respond(c, this.svc.handleBachsEvent(event), "webhook received")
}

func (this *httpHandler) completeCheckout(c *gin.Context) {
	net.Respond(c, this.svc.completeCheckout(c.Param("reference")), "checkout completed")
}
