package payment

import "github.com/gin-gonic/gin"

type router struct {
	http *httpHandler
	path string
}

func CreateRouter(http *httpHandler, path string) *router {
	return &router{
		http: http,
		path: path,
	}
}

func (this *router) GetPath() string {
	return this.path
}

func (this *router) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/", this.http.createPayment)
	router.GET("/:id", this.http.getPaymentInfo)
	router.GET("/", this.http.getPayments)
	router.GET("/stats", this.http.getTotalPaymentStats)
	router.POST("/checkout", this.http.initiateCheckout)
}
