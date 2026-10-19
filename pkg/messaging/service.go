package messaging

import "github.com/patrickishaf/property-mgt/common"

type service struct{}

func CreateService() *service {
	return &service{}
}

func (this *service) getMessageRecipients() common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) getMessageInfo() common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) getMessages() common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) getTemplates() common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) sendMessage() common.ServiceResponse {
	return common.ServiceResponse{}
}
