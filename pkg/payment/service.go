package payment

import "github.com/patrickishaf/property-mgt/common"

type service struct{}

func CreateService() *service {
	return &service{}
}

func (this *service) createPayment() common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) getPaymentInfo() common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) getPayments() common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) getTotalPaymentStats() common.ServiceResponse {
	return common.ServiceResponse{}
}

func (this *service) initiateCheckout() common.ServiceResponse {
	return common.ServiceResponse{}
}
