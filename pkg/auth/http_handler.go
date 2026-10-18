package auth

import (
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
	if err := c.ShouldBindJSON(body); err != nil {
		c.IndentedJSON(400, common.ErrInvalidRequest)
		return
	}

	res := this.srv.createOwnerAccount(c.Request.Context(), body)
	if res.Err != nil {
		c.IndentedJSON(res.StatusCode, res.Err.Error())
	} else {
		c.IndentedJSON(res.StatusCode, res.Data)
	}
}

func (this *httpHandler) getAuthenticatedUser(c *gin.Context) {
	res := this.srv.getAuthenticatedUser(c.Request.Context())
	if res.Err != nil {
		c.IndentedJSON(res.StatusCode, res.Err.Error())
	} else {
		c.IndentedJSON(res.StatusCode, res.Data)
	}
}

func (this *httpHandler) loginWithPassword(c *gin.Context) {
	res := this.srv.loginWithPassword(c.Request.Context())
	if res.Err != nil {
		c.IndentedJSON(res.StatusCode, res.Err.Error())
	} else {
		c.IndentedJSON(res.StatusCode, res.Data)
	}
}

func (this *httpHandler) logout(c *gin.Context) {
	res := this.srv.logout(c.Request.Context())
	if res.Err != nil {
		c.IndentedJSON(res.StatusCode, res.Err.Error())
	} else {
		c.IndentedJSON(res.StatusCode, res.Data)
	}
}

func (this *httpHandler) sendOTP(c *gin.Context) {
	res := this.srv.sendOTP(c.Request.Context())
	if res.Err != nil {
		c.IndentedJSON(res.StatusCode, res.Err.Error())
	} else {
		c.IndentedJSON(res.StatusCode, res.Data)
	}
}

func (this *httpHandler) verifyOTP(c *gin.Context) {
	res := this.srv.verifyOTP(c.Request.Context())
	if res.Err != nil {
		c.IndentedJSON(res.StatusCode, res.Err.Error())
	} else {
		c.IndentedJSON(res.StatusCode, res.Data)
	}
}
