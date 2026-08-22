package auth

import "github.com/gin-gonic/gin"

type httpHandler struct {
	srv *service
}

func CreateHTTPHandler(srv *service) *httpHandler {
	return &httpHandler{
		srv: srv,
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
