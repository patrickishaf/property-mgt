package maintenance

import (
	"github.com/gin-gonic/gin"
	"github.com/patrickishaf/property-mgt/config"
	"github.com/patrickishaf/property-mgt/internal/net"
)

type router struct {
	cfg  *config.AppConfig
	http *httpHandler
	path string
}

func CreateRouter(cfg *config.AppConfig, http *httpHandler, path string) *router {
	return &router{
		cfg:  cfg,
		http: http,
		path: path,
	}
}

func (this *router) GetPath() string {
	return this.path
}

func (this *router) RegisterRoutes(router *gin.RouterGroup) {
	router.Use(net.VerifyJWT(this.cfg.JWTSecretKey))

	router.POST("/complaints", this.http.createComplaint)
	router.GET("/complaints", this.http.getComplaints)
	router.GET("/complaints/linked", this.http.getLinkedComplaints)
	router.GET("/complaints/:id", this.http.getComplaintInfo)
	router.GET("/complaints/:id/updates", this.http.getComplaintUpdates)

	router.POST("/issues", this.http.createIssue)
	router.GET("/issues", this.http.getIssues)
	router.GET("/issues/:id", this.http.getIssueInfo)
	router.GET("/issues/:id/updates", this.http.getIssueUpdates)
	router.POST("/issues/:id/updates", this.http.postIssueUpdate)
	router.POST("/issues/:id/complaints", this.http.linkComplaints)
}
