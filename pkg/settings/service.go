package settings

import "github.com/patrickishaf/property-mgt/common"

type service struct{}

func NewService() *service {
	return &service{}
}

func (this *service) getUserSettings() common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) updateProfileSettings() common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) updatePreferences() common.ServiceResponse {
	return common.ServiceResponse{}
}
