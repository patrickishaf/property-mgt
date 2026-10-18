package auth

import (
	"context"
	"errors"

	"github.com/patrickishaf/property-mgt/common"
)

type service struct{}

func CreateService() *service {
	return &service{}
}

func (this *service) createOwnerAccount(ctx context.Context, dto createOwnerAccountReq) common.ServiceResponse {
	return common.ServiceResponse{StatusCode: 200, Err: errors.New(common.ErrMethodNotImplemented)}
}

func (this *service) getAuthenticatedUser(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{StatusCode: 200, Err: errors.New(common.ErrMethodNotImplemented)}
}

func (this *service) loginWithPassword(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{StatusCode: 200, Err: errors.New(common.ErrMethodNotImplemented)}
}

func (this *service) logout(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{StatusCode: 200, Err: errors.New(common.ErrMethodNotImplemented)}
}

func (this *service) refreshAuthToken(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{StatusCode: 200, Err: errors.New(common.ErrMethodNotImplemented)}
}

func (this *service) sendOTP(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{StatusCode: 401, Err: errors.New(common.ErrTokenExpired)}
}

func (this *service) validateAccessToken(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{StatusCode: 200, Err: errors.New(common.ErrMethodNotImplemented)}
}

func (this *service) verifyOTP(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{StatusCode: 200, Err: errors.New(common.ErrMethodNotImplemented)}
}
