package maintenance

import "github.com/patrickishaf/property-mgt/common"

type service struct{}

func NewService() *service {
	return &service{}
}

func (this *service) createComplaint() common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) createIssue() common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) getComplaints() common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) getIssues() common.ServiceResponse {
	return common.ServiceResponse{}
}
