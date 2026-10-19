package payment

import "github.com/gin-gonic/gin"

type httpHandler struct {
	svc *service
}

func CreateHTTPHandler(svc *service) *httpHandler {
	return &httpHandler{
		svc: svc,
	}
}

func (this *httpHandler) createPayment(c *gin.Context) {}

func (this *httpHandler) getPaymentInfo(c *gin.Context) {}

func (this *httpHandler) getPayments(c *gin.Context) {}

func (this *httpHandler) getTotalPaymentStats(c *gin.Context) {}

func (this *httpHandler) initiateCheckout(c *gin.Context) {}
