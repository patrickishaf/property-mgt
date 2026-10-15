package messaging

import "github.com/patrickishaf/property-mgt/common"

type service struct{}

func NewService() *service {
	return &service{}
}

func (this *service) getMessages() common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) sendMessage() common.ServiceResponse {
	return common.ServiceResponse{}
}
