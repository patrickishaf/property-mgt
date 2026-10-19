package maintenance

import "github.com/gin-gonic/gin"

type httpHandler struct {
	svc *service
}

func CreateHTTPHandler(svc *service) *httpHandler {
	return &httpHandler{
		svc: svc,
	}
}

func (this *httpHandler) createComplaint(c *gin.Context) {}

func (this *httpHandler) createIssue(c *gin.Context) {}

func (this *httpHandler) getComplaintInfo(c *gin.Context) {}

func (this *httpHandler) getComplaints(c *gin.Context) {}

func (this *httpHandler) getComplaintUpdates(c *gin.Context) {}

func (this *httpHandler) getIssueInfo(c *gin.Context) {}

func (this *httpHandler) getIssues(c *gin.Context) {}

func (this *httpHandler) getIssueUpdates(c *gin.Context) {}

func (this *httpHandler) getLinkedComplaints(c *gin.Context) {}
