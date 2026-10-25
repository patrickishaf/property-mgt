package auth

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/patrickishaf/property-mgt/common"
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
		c.IndentedJSON(400, common.ErrInvalidRequest)
		return
	}

	res := this.srv.createOwnerAccount(body)
	if res.Err != "" {
		c.IndentedJSON(res.StatusCode, res.Err)
	} else {
		c.IndentedJSON(res.StatusCode, res.Data)
	}
}

func (this *httpHandler) getAuthenticatedUser(c *gin.Context) {
	value, exists := c.Get(common.CTX_KEY_ACTOR)
	if !exists {
		log.Println("no actor in context")
		c.IndentedJSON(400, common.ErrInvalidCredentials)
		return
	}
	actor, isValid := value.(*common.Actor)
	if !isValid {
		log.Println("invalid actor in context")
		c.IndentedJSON(400, common.ErrInvalidCredentials)
		return
	}
	res := this.srv.getAuthenticatedUser(actor)
	if res.Err != "" {
		c.IndentedJSON(res.StatusCode, res.Err)
	} else {
		c.IndentedJSON(res.StatusCode, res.Data)
	}
}

func (this *httpHandler) loginWithPassword(c *gin.Context) {
	res := this.srv.loginWithPassword(passwordLoginReq{})
	if res.Err != "" {
		c.IndentedJSON(res.StatusCode, res.Err)
	} else {
		c.IndentedJSON(res.StatusCode, res.Data)
	}
}

func (this *httpHandler) verifyOTP(c *gin.Context) {
	res := this.srv.verifyOTP(c.Request.Context())
	if res.Err != "" {
		c.IndentedJSON(res.StatusCode, res.Err)
	} else {
		c.IndentedJSON(res.StatusCode, res.Data)
	}
}
