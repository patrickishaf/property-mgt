package property

import "github.com/patrickishaf/property-mgt/common"

type service struct{}

func NewService() *service {
	return &service{}
}

func (this *service) createPropertyManager() common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) getProperties() common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) getPropertyInfo() common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) getPropertyManager() common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) getPropertyManagers() common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) removePropertyManager() common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) updatePropertyInfo() common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) updatePropertyManager() common.ServiceResponse {
	return common.ServiceResponse{}
}
