package messaging

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

func requestActor(c *gin.Context) (*common.Actor, bool) {
	actor, ok := net.GetActor(c)
	if !ok {
		c.IndentedJSON(401, net.ErrorResponse(common.ErrInvalidCredentials, nil))
	}
	return actor, ok
}

func (this *httpHandler) getMessageRecipients(c *gin.Context) {
	actor, ok := requestActor(c)
	if !ok {
		return
	}
	var query getRecipientsReq
	if err := c.ShouldBindQuery(&query); err != nil {
		log.Println("failed to bind query", "err", err.Error())
		c.IndentedJSON(400, net.ErrorResponse(common.ErrInvalidRequest, nil))
		return
	}
	net.Respond(c, this.svc.getMessageRecipients(actor, query), "")
}

func (this *httpHandler) getMessageInfo(c *gin.Context) {
	actor, ok := requestActor(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.IndentedJSON(400, net.ErrorResponse(common.ErrInvalidRequest, nil))
		return
	}
	net.Respond(c, this.svc.getMessageInfo(actor, uint(id)), "")
}

func (this *httpHandler) getMessages(c *gin.Context) {
	actor, ok := requestActor(c)
	if !ok {
		return
	}
	var query getMessagesReq
	if err := c.ShouldBindQuery(&query); err != nil {
		log.Println("failed to bind query", "err", err.Error())
		c.IndentedJSON(400, net.ErrorResponse(common.ErrInvalidRequest, nil))
		return
	}
	net.Respond(c, this.svc.getMessages(actor, query), "")
}

func (this *httpHandler) getTemplates(c *gin.Context) {
	actor, ok := requestActor(c)
	if !ok {
		return
	}
	net.Respond(c, this.svc.getTemplates(actor), "")
}

func (this *httpHandler) sendMessage(c *gin.Context) {
	actor, ok := requestActor(c)
	if !ok {
		return
	}
	var body sendMessageReq
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Println("failed to bind json", "err", err.Error())
		c.IndentedJSON(400, net.ErrorResponse(common.ErrInvalidRequest, nil))
		return
	}
	net.Respond(c, this.svc.sendMessage(actor, body), "message sent")
}
