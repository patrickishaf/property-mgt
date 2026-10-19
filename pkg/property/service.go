package property

import (
	"context"

	"github.com/patrickishaf/property-mgt/common"
)

type service struct{}

func CreateService() *service {
	return &service{}
}

func (this *service) createPropertyManager(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) getProperties(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) getPropertyInfo(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) getPropertyManager(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) getPropertyManagers(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) removePropertyManager(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) updatePropertyInfo(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) updatePropertyManager(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{}
}
