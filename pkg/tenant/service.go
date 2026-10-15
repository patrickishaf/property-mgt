package tenant

import "github.com/patrickishaf/property-mgt/common"

type service struct{}

func NewService() *service {
	return &service{}
}

func (this *service) createPaymentForTenant() common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) createRentObligation() common.ServiceResponse {
	return common.ServiceResponse{}
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

func (this *service) updateTenantInfo() common.ServiceResponse {
	return common.ServiceResponse{}
}
