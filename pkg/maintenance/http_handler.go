package maintenance

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

// requestActor returns the authenticated actor and the requested numeric path
// ids, writing an error response and returning ok=false if any is invalid.
func requestActor(c *gin.Context, params ...string) (*common.Actor, []uint, bool) {
	actor, ok := net.GetActor(c)
	if !ok {
		c.IndentedJSON(401, net.ErrorResponse(common.ErrInvalidCredentials, nil))
		return nil, nil, false
	}
	ids := make([]uint, 0, len(params))
	for _, name := range params {
		id, err := strconv.ParseUint(c.Param(name), 10, 64)
		if err != nil || id == 0 {
			c.IndentedJSON(400, net.ErrorResponse(common.ErrInvalidRequest, nil))
			return nil, nil, false
		}
		ids = append(ids, uint(id))
	}
	return actor, ids, true
}

func bindJSON(c *gin.Context, body any) bool {
	if err := c.ShouldBindJSON(body); err != nil {
		log.Println("failed to bind json", "err", err.Error())
		c.IndentedJSON(400, net.ErrorResponse(common.ErrInvalidRequest, nil))
		return false
	}
	return true
}

func bindQuery(c *gin.Context, query any) bool {
	if err := c.ShouldBindQuery(query); err != nil {
		log.Println("failed to bind query", "err", err.Error())
		c.IndentedJSON(400, net.ErrorResponse(common.ErrInvalidRequest, nil))
		return false
	}
	return true
}

func (this *httpHandler) createComplaint(c *gin.Context) {
	actor, _, ok := requestActor(c)
	var body createComplaintReq
	if !ok || !bindJSON(c, &body) {
		return
	}
	net.Respond(c, this.svc.createComplaint(actor, body), "complaint created")
}

func (this *httpHandler) createIssue(c *gin.Context) {
	actor, _, ok := requestActor(c)
	var body createIssueReq
	if !ok || !bindJSON(c, &body) {
		return
	}
	net.Respond(c, this.svc.createIssue(actor, body), "issue created")
}

func (this *httpHandler) getComplaintInfo(c *gin.Context) {
	actor, ids, ok := requestActor(c, "id")
	if !ok {
		return
	}
	net.Respond(c, this.svc.getComplaintInfo(actor, ids[0]), "")
}

func (this *httpHandler) getComplaints(c *gin.Context) {
	actor, _, ok := requestActor(c)
	var query getComplaintsReq
	if !ok || !bindQuery(c, &query) {
		return
	}
	net.Respond(c, this.svc.getComplaints(actor, query), "")
}

func (this *httpHandler) getComplaintUpdates(c *gin.Context) {
	actor, ids, ok := requestActor(c, "id")
	if !ok {
		return
	}
	net.Respond(c, this.svc.getComplaintUpdates(actor, ids[0]), "")
}

func (this *httpHandler) getIssueInfo(c *gin.Context) {
	actor, ids, ok := requestActor(c, "id")
	if !ok {
		return
	}
	net.Respond(c, this.svc.getIssueInfo(actor, ids[0]), "")
}

func (this *httpHandler) getIssues(c *gin.Context) {
	actor, _, ok := requestActor(c)
	var query getIssuesReq
	if !ok || !bindQuery(c, &query) {
		return
	}
	net.Respond(c, this.svc.getIssues(actor, query), "")
}

func (this *httpHandler) getIssueUpdates(c *gin.Context) {
	actor, ids, ok := requestActor(c, "id")
	if !ok {
		return
	}
	net.Respond(c, this.svc.getIssueUpdates(actor, ids[0]), "")
}

func (this *httpHandler) getLinkedComplaints(c *gin.Context) {
	actor, _, ok := requestActor(c)
	var query getLinkedComplaintsReq
	if !ok || !bindQuery(c, &query) {
		return
	}
	net.Respond(c, this.svc.getLinkedComplaints(actor, query), "")
}

func (this *httpHandler) linkComplaints(c *gin.Context) {
	actor, ids, ok := requestActor(c, "id")
	var body linkComplaintsReq
	if !ok || !bindJSON(c, &body) {
		return
	}
	net.Respond(c, this.svc.linkComplaints(actor, ids[0], body), "complaints linked")
}

func (this *httpHandler) postIssueUpdate(c *gin.Context) {
	actor, ids, ok := requestActor(c, "id")
	var body postIssueUpdateReq
	if !ok || !bindJSON(c, &body) {
		return
	}
	net.Respond(c, this.svc.postIssueUpdate(actor, ids[0], body), "issue updated")
}
