package maintenance

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
	router.POST("/complaints", this.http.createComplaint)
	router.POST("/issues", this.http.createIssue)
	router.GET("/complaints/:id", this.http.getComplaintInfo)
	router.GET("/complaints", this.http.getComplaints)
	router.GET("/complaints/:id/updates", this.http.getComplaintUpdates)
	router.GET("/issues", this.http.getIssues)
	router.GET("/issues/:id", this.http.getIssueInfo)
	router.GET("/issues/:id/updates", this.http.getIssueUpdates)
	router.GET("/complaints/linked", this.http.getLinkedComplaints)
}
