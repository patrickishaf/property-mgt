package auth

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/internal/net"
)

type httpHandler struct {
	srv *service
}

func CreateHTTPHandler(srv *service) *httpHandler {
	return &httpHandler{
		srv: srv,
	}
}

func (this *httpHandler) createOwnerAccount(c *gin.Context) {
	var body createOwnerAccountReq
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Println("failed to bind json", "err", err.Error())
		c.IndentedJSON(400, net.ErrorResponse(common.ErrInvalidRequest, nil))
		return
	}
	net.Respond(c, this.srv.createOwnerAccount(body), "account created")
}

func (this *httpHandler) getAuthenticatedUser(c *gin.Context) {
	actor, ok := net.GetActor(c)
	if !ok {
		log.Println("no valid actor in context")
		c.IndentedJSON(401, net.ErrorResponse(common.ErrInvalidCredentials, nil))
		return
	}
	net.Respond(c, this.srv.getAuthenticatedUser(actor), "")
}

func (this *httpHandler) loginWithPassword(c *gin.Context) {
	var body passwordLoginReq
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Println("failed to bind json", "err", err.Error())
		c.IndentedJSON(400, net.ErrorResponse(common.ErrInvalidRequest, nil))
		return
	}
	net.Respond(c, this.srv.loginWithPassword(body), "login successful")
}

func (this *httpHandler) refreshAuthToken(c *gin.Context) {
	var body refreshAuthTokenReq
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Println("failed to bind json", "err", err.Error())
		c.IndentedJSON(400, net.ErrorResponse(common.ErrInvalidRequest, nil))
		return
	}
	net.Respond(c, this.srv.refreshAuthToken(body), "token refreshed")
}

func (this *httpHandler) setPassword(c *gin.Context) {
	var body setPasswordReq
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Println("failed to bind json", "err", err.Error())
		c.IndentedJSON(400, net.ErrorResponse(common.ErrInvalidRequest, nil))
		return
	}
	net.Respond(c, this.srv.setPassword(body), "password set")
}
