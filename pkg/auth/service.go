package auth

import (
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
	adminRepo  *repository.AdminRepository
	cfg        *config.AppConfig
	logger     *slog.Logger
	tenantRepo *repository.TenantRepository
}

func CreateService(cfg *config.AppConfig, logger *slog.Logger, adminRepo *repository.AdminRepository, tenantRepo *repository.TenantRepository,
) *service {
	return &service{
		adminRepo:  adminRepo,
		cfg:        cfg,
		logger:     logger,
		tenantRepo: tenantRepo,
	}
}

func (this *service) createOwnerAccount(dto createOwnerAccountReq) common.ServiceResponse {
	logger := this.logger.With("service", "auth.service", "caller", "auth.service", "actor", dto.Email)
	/**
	- find an existing owner with that email
	- if owner exists, return error
	*/
	admin, err := this.adminRepo.FindOneByEmailAndRole(dto.Email, common.AdminRoleOwner)
	if err != nil {
		logger.Error("failed to find existing owner by email", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	if admin != nil {
		logger.Info("admin is nil", "admin", fmt.Sprintf("%+v", admin))
		return common.ServiceResponse{StatusCode: 409, Err: "email already taken"}
	}

	pwdHash, err := common.HashPassword(dto.Password)
	if err != nil {
		logger.Error("failed to hash password")
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
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
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}

	admin, err = this.adminRepo.FindOneByEmailAndRole(dto.Email, common.AdminRoleOwner)
	actor := common.Actor{
		ID:          admin.ID,
		Email:       admin.Email,
		FirstName:   admin.FirstName,
		LastName:    admin.LastName,
		PhoneNumber: admin.PhoneNumber,
		Role:        admin.Role,
	}
	accessToken, err := common.GenerateJWT(*this.cfg, actor)
	if err != nil {
		logger.Error("failed to generate access token", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	refreshToken, err := this.generateRefreshToken(actor)
	if err != nil {
		logger.Error("failed to generate access token", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	return common.ServiceResponse{StatusCode: 200, Data: map[string]any{
		"user":         admin.Safe(),
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	}}
}

func (this *service) generateRefreshToken(actor common.Actor) (string, error) {
	expiryDate := time.Now().Add(time.Hour * time.Duration(this.cfg.RefreshTokenExpirationHours))
	payload := common.Claims{
		Actor:     actor,
		TokenType: common.TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiryDate),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &payload)
	return token.SignedString([]byte(this.cfg.JWTSecretKey))
}

func (this *service) getAuthenticatedUser(actor *common.Actor) common.ServiceResponse {
	logger := this.logger.With("service", "auth.service", "caller", "auth.service.getAuthenticatedUser", "actor", actor.Email)

	if actor == nil {
		return common.ServiceResponse{StatusCode: 401, Err: common.ErrInvalidCredentials}
	}

	if actor.Role == common.AdminRoleOwner.ToString() || actor.Role == common.AdminRoleManager.ToString() {
		admin, err := this.adminRepo.FindOneById(actor.ID)
		if err != nil {
			logger.Error("failed to find admin by id", "err", err.Error())
			return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
		}
		if admin == nil {
			return common.ServiceResponse{StatusCode: 404, Err: common.ErrDBRecordNotFound}
		}
		return common.ServiceResponse{StatusCode: 200, Data: admin.Safe()}
	} else if actor.Role == common.ActorRoleTenant {
		tenant, err := this.tenantRepo.FindOneById(actor.ID)
		if err != nil {
			logger.Error("failed to find tenant by id", "err", err.Error())
			return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
		}
		if tenant == nil {
			return common.ServiceResponse{StatusCode: 404, Err: common.ErrDBRecordNotFound}
		}
		return common.ServiceResponse{StatusCode: 200, Data: tenant.SafeMap()}
	}
	logger.Error("actor role is unrecognized", "actor_role", actor.Role)
	return common.ServiceResponse{StatusCode: 400, Err: common.ErrInvalidRequest}
}

func (this *service) loginWithPassword(dto passwordLoginReq) common.ServiceResponse {
	logger := this.logger.With("caller", "auth.service.loginWithPassword", "actor", dto.Email)

	if dto.Type == "tenant" {
		tenant, err := this.tenantRepo.FindOneByEmail(dto.Email)
		if err != nil {
			logger.Error("failed to find tenant by email", "err", err.Error())
			return common.ServiceResponse{StatusCode: 401, Err: common.ErrDBRecordNotFound}
		}

		if tenant == nil || tenant.PasswordHash == "" {
			return common.ServiceResponse{StatusCode: 401, Err: common.ErrInvalidCredentials}
		}

		yes, err := common.DoPasswordsMatch(dto.Password, tenant.PasswordHash)
		if err != nil {
			logger.Error("failed to compare passwords", "err", err.Error())
			return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
		}
		if !yes {
			logger.Error("passwords do not match")
			return common.ServiceResponse{StatusCode: 401, Err: common.ErrInvalidCredentials}
		}

		actor := common.Actor{
			ID:          tenant.ID,
			Email:       tenant.Email,
			FirstName:   tenant.FirstName,
			LastName:    tenant.LastName,
			PhoneNumber: tenant.PhoneNumber,
			Role:        common.ActorRoleTenant,
		}
		accessToken, err := common.GenerateJWT(*this.cfg, actor)
		if err != nil {
			logger.Error("failed to generate access token", "err", err.Error())
			return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
		}
		refreshToken, err := this.generateRefreshToken(actor)
		if err != nil {
			logger.Error("failed to generate access token", "err", err.Error())
			return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
		}

		return common.ServiceResponse{StatusCode: 200, Data: map[string]any{
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
		}}
	} else {
		admin, err := this.adminRepo.FindOneByEmail(dto.Email)
		if err != nil {
			logger.Error("failed to find admin by email", "err", err.Error())
			return common.ServiceResponse{StatusCode: 401, Err: common.ErrDBRecordNotFound}
		}

		if admin == nil || admin.PasswordHash == "" {
			return common.ServiceResponse{StatusCode: 401, Err: common.ErrInvalidCredentials}
		}

		yes, err := common.DoPasswordsMatch(dto.Password, admin.PasswordHash)
		if err != nil {
			logger.Error("failed to compare passwords", "err", err.Error())
			return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
		}
		if !yes {
			logger.Error("passwords do not match")
			return common.ServiceResponse{StatusCode: 401, Err: common.ErrInvalidCredentials}
		}

		actor := common.Actor{
			ID:          admin.ID,
			Email:       admin.Email,
			FirstName:   admin.FirstName,
			LastName:    admin.LastName,
			PhoneNumber: admin.PhoneNumber,
			Role:        admin.Role,
		}
		accessToken, err := common.GenerateJWT(*this.cfg, actor)
		if err != nil {
			logger.Error("failed to generate access token", "err", err.Error())
			return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
		}
		refreshToken, err := this.generateRefreshToken(actor)
		if err != nil {
			logger.Error("failed to generate access token", "err", err.Error())
			return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
		}

		return common.ServiceResponse{StatusCode: 200, Data: map[string]any{
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
		}}
	}
}

func (this *service) refreshAuthToken(dto refreshAuthTokenReq) common.ServiceResponse {
	logger := this.logger.With("caller", "auth.service.refreshAuthToken")
	jwtKey := []byte(this.cfg.JWTSecretKey)

	refreshClaims, err := common.ParseJWT(dto.RefreshToken, jwtKey, false)
	if err != nil || refreshClaims.TokenType != common.TokenTypeRefresh {
		logger.Error("invalid refresh token")
		return common.ServiceResponse{StatusCode: 401, Err: common.ErrInvalidToken}
	}
	accessClaims, err := common.ParseJWT(dto.AccessToken, jwtKey, true)
	if err != nil || accessClaims.TokenType != common.TokenTypeAccess {
		logger.Error("invalid access token")
		return common.ServiceResponse{StatusCode: 401, Err: common.ErrInvalidToken}
	}
	if refreshClaims.Actor.ID != accessClaims.Actor.ID || refreshClaims.Actor.Role != accessClaims.Actor.Role {
		logger.Error("access and refresh tokens belong to different actors")
		return common.ServiceResponse{StatusCode: 401, Err: common.ErrInvalidToken}
	}

	claimedActor := refreshClaims.Actor
	logger = logger.With("actor", claimedActor.Email)
	var actor common.Actor
	if claimedActor.Role == common.AdminRoleOwner.ToString() || claimedActor.Role == common.AdminRoleManager.ToString() {
		admin, err := this.adminRepo.FindOneById(claimedActor.ID)
		if err != nil {
			logger.Error("failed to find admin by id", "err", err.Error())
			return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
		}
		if admin == nil {
			return common.ServiceResponse{StatusCode: 401, Err: common.ErrInvalidCredentials}
		}
		actor = common.Actor{
			ID:          admin.ID,
			Email:       admin.Email,
			FirstName:   admin.FirstName,
			LastName:    admin.LastName,
			PhoneNumber: admin.PhoneNumber,
			Role:        admin.Role,
		}
	} else if claimedActor.Role == common.ActorRoleTenant {
		tenant, err := this.tenantRepo.FindOneById(claimedActor.ID)
		if err != nil {
			logger.Error("failed to find tenant by id", "err", err.Error())
			return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
		}
		if tenant == nil {
			return common.ServiceResponse{StatusCode: 401, Err: common.ErrInvalidCredentials}
		}
		actor = common.Actor{
			ID:          tenant.ID,
			Email:       tenant.Email,
			FirstName:   tenant.FirstName,
			LastName:    tenant.LastName,
			PhoneNumber: tenant.PhoneNumber,
			Role:        common.ActorRoleTenant,
		}
	} else {
		logger.Error("actor role is unrecognized", "actor_role", claimedActor.Role)
		return common.ServiceResponse{StatusCode: 401, Err: common.ErrInvalidToken}
	}

	accessToken, err := common.GenerateJWT(*this.cfg, actor)
	if err != nil {
		logger.Error("failed to generate access token", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	return common.ServiceResponse{StatusCode: 200, Data: map[string]any{
		"accessToken": accessToken,
	}}
}

func (this *service) setPassword(dto setPasswordReq) common.ServiceResponse {
	logger := this.logger.With("caller", "auth.service.setPassword")

	claims, err := common.ParseJWT(dto.Token, []byte(this.cfg.JWTSecretKey), false)
	if err != nil || claims.TokenType != common.TokenTypeInvite {
		logger.Error("invalid invite token")
		return common.ServiceResponse{StatusCode: 401, Err: common.ErrInvalidToken}
	}

	var actor common.Actor
	var passwordAlreadySet bool
	var savePasswordHash func(hash string) error
	switch claims.Actor.Role {
	case common.ActorRoleTenant:
		tenant, err := this.tenantRepo.FindOneById(claims.Actor.ID)
		if err != nil {
			logger.Error("failed to find tenant by id", "err", err.Error())
			return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
		}
		if tenant == nil {
			return common.ServiceResponse{StatusCode: 401, Err: common.ErrInvalidToken}
		}
		actor = common.Actor{
			ID:          tenant.ID,
			Email:       tenant.Email,
			FirstName:   tenant.FirstName,
			LastName:    tenant.LastName,
			PhoneNumber: tenant.PhoneNumber,
			Role:        common.ActorRoleTenant,
		}
		passwordAlreadySet = tenant.PasswordHash != ""
		savePasswordHash = func(hash string) error { return this.tenantRepo.UpdatePasswordHash(tenant.ID, hash) }
	case common.AdminRoleManager.ToString():
		admin, err := this.adminRepo.FindOneById(claims.Actor.ID)
		if err != nil {
			logger.Error("failed to find admin by id", "err", err.Error())
			return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
		}
		if admin == nil || admin.Role != common.AdminRoleManager.ToString() {
			return common.ServiceResponse{StatusCode: 401, Err: common.ErrInvalidToken}
		}
		actor = common.Actor{
			ID:          admin.ID,
			Email:       admin.Email,
			FirstName:   admin.FirstName,
			LastName:    admin.LastName,
			PhoneNumber: admin.PhoneNumber,
			Role:        admin.Role,
		}
		passwordAlreadySet = admin.PasswordHash != ""
		savePasswordHash = func(hash string) error { return this.adminRepo.UpdatePasswordHash(admin.ID, hash) }
	default:
		logger.Error("invite token has unsupported role", "role", claims.Actor.Role)
		return common.ServiceResponse{StatusCode: 401, Err: common.ErrInvalidToken}
	}

	logger = logger.With("actor", actor.Email)
	if passwordAlreadySet {
		return common.ServiceResponse{StatusCode: 409, Err: "password already set"}
	}

	pwdHash, err := common.HashPassword(dto.Password)
	if err != nil {
		logger.Error("failed to hash password", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	if err := savePasswordHash(pwdHash); err != nil {
		logger.Error("failed to save password hash", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}

	accessToken, err := common.GenerateJWT(*this.cfg, actor)
	if err != nil {
		logger.Error("failed to generate access token", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	refreshToken, err := this.generateRefreshToken(actor)
	if err != nil {
		logger.Error("failed to generate refresh token", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	return common.ServiceResponse{StatusCode: 200, Data: map[string]any{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	}}
}
