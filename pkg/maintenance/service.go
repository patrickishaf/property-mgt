package maintenance

import (
	"context"

	"github.com/patrickishaf/property-mgt/common"
)

type service struct{}

func CreateService() *service {
	return &service{}
}

func (this *service) createComplaint(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) createIssue(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) getComplaintInfo(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) getComplaints(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) getComplaintUpdates(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) getIssueInfo(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) getIssues(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) getIssueUpdates(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) getLinkedComplaints(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{}
}
