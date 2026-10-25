package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/config"
	"github.com/patrickishaf/property-mgt/internal/db/model"
	"github.com/patrickishaf/property-mgt/internal/db/repository"
)

type service struct {
	adminRepo *repository.AdminRepository
	cfg       *config.AppConfig
	logger    *slog.Logger
}

func CreateService(cfg *config.AppConfig, logger *slog.Logger, adminRepo *repository.AdminRepository) *service {
	return &service{
		adminRepo: adminRepo,
		cfg:       cfg,
		logger:    logger,
	}
}

func (this *service) createOwnerAccount(ctx context.Context, dto createOwnerAccountReq) common.ServiceResponse {
	logger := this.logger.With("service", "auth.service", "caller", "auth.service", "actor", dto.Email)
	/**
	- find an existing owner with that email
	- if owner exists, return error
	*/
	admin, err := this.adminRepo.FindOneByEmail(dto.Email, common.AdminRoleOwner)
	if err != nil {
		logger.Error("failed to find existing owner by email", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: errors.New(common.ErrGeneric)}
	}
	if admin != nil {
		logger.Info("admin is nil", "admin", fmt.Sprintf("%+v", admin))
		return common.ServiceResponse{StatusCode: 401, Err: errors.New("email already taken")}
	}

	pwdHash, err := common.HashPassword(dto.Password)
	if err != nil {
		logger.Error("failed to hash password")
		return common.ServiceResponse{StatusCode: 500, Err: errors.New(common.ErrGeneric)}
	}
	err = this.adminRepo.SaveOne(&model.Admin{
		Email:        dto.Email,
		FirstName:    dto.FirstName,
		LastName:     dto.LastName,
		PasswordHash: pwdHash,
		PhoneNumber:  dto.PhoneNumber,
		Role:         common.AdminRoleOwner.ToString(),
	})
	if err != nil {
		logger.Error("failed to save admin to repo", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: errors.New(common.ErrGeneric)}
	}

	/*
		sign in this user and get the
	*/
	admin, err = this.adminRepo.FindOneByEmail(dto.Email, common.AdminRoleOwner)
	actor := common.Actor{
		ID:          admin.ID,
		Email:       admin.Email,
		FirstName:   admin.FirstName,
		LastName:    admin.LastName,
		PhoneNumber: admin.PhoneNumber,
		Role:        admin.Role,
	}
	accessToken, err := this.generateAccessToken(actor)
	if err != nil {
		logger.Error("failed to generate access token", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: errors.New(common.ErrGeneric)}
	}
	refreshToken, err := this.generateRefreshToken(actor)
	if err != nil {
		logger.Error("failed to generate access token", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: errors.New(common.ErrGeneric)}
	}
	return common.ServiceResponse{StatusCode: 200, Data: map[string]any{
		"user": map[string]string{
			"email":       actor.Email,
			"firstName":   actor.FirstName,
			"lastName":    actor.LastName,
			"phoneNumber": actor.PhoneNumber,
		},
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	}}
}

func (this *service) generateAccessToken(actor common.Actor) (string, error) {
	expiryDate := time.Now().Add(time.Minute * time.Duration(this.cfg.AccessTokenExpirationMinutes))
	claims := jwtPayload{
		Actor: actor,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiryDate),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	return token.SignedString([]byte(this.cfg.JWTSecretKey))
}

func (this *service) generateRefreshToken(actor common.Actor) (string, error) {
	expiryDate := time.Now().Add(time.Hour * time.Duration(this.cfg.RefreshTokenExpirationHours))
	payload := jwtPayload{
		Actor: actor,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiryDate),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &payload)
	return token.SignedString([]byte(this.cfg.JWTSecretKey))
}

func (this *service) getAuthenticatedUser(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{StatusCode: 200, Err: errors.New(common.ErrMethodNotImplemented)}
}

func (this *service) loginWithPassword(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{StatusCode: 200, Err: errors.New(common.ErrMethodNotImplemented)}
}

func (this *service) refreshAuthToken(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{StatusCode: 200, Err: errors.New(common.ErrMethodNotImplemented)}
}

func (this *service) validateAccessToken(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{StatusCode: 200, Err: errors.New(common.ErrMethodNotImplemented)}
}

func (this *service) verifyOTP(ctx context.Context) common.ServiceResponse {
	return common.ServiceResponse{StatusCode: 200, Err: errors.New(common.ErrMethodNotImplemented)}
}
