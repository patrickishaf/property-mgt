package tenant

import "github.com/patrickishaf/property-mgt/common"

type service struct{}

func CreateService() *service {
	return &service{}
}

func (this *service) createTenant() common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) getTenants() common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) getTenantInfo() common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) sendDirectMessage() common.ServiceResponse {
	return common.ServiceResponse{}
}
