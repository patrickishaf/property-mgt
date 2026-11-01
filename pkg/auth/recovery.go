package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/providers/cpaas"
)

// passwordVersion fingerprints a password hash for reset tokens. A token only
// works while the account still has the password it was issued against, so it
// can't be reused once the password has been reset (or changed some other way).
func passwordVersion(passwordHash string) string {
	sum := sha256.Sum256([]byte(passwordHash))
	return hex.EncodeToString(sum[:8])
}

// recoveryAccount is the account a recovery request or token refers to, with
// what's needed to email it, check its token and save its new password.
type recoveryAccount struct {
	actor        common.Actor
	passwordHash string
	savePassword func(hash string) error
}

func (this *service) findRecoveryAccountByEmail(email string, accountType string) (*recoveryAccount, error) {
	if accountType == "tenant" {
		tenant, err := this.tenantRepo.FindOneByEmail(email)
		if err != nil || tenant == nil {
			return nil, err
		}
		return &recoveryAccount{
			actor: common.Actor{ID: tenant.ID, Email: tenant.Email, FirstName: tenant.FirstName,
				LastName: tenant.LastName, PhoneNumber: tenant.PhoneNumber, Role: common.ActorRoleTenant},
			passwordHash: tenant.PasswordHash,
			savePassword: func(hash string) error { return this.tenantRepo.UpdatePasswordHash(tenant.ID, hash) },
		}, nil
	}
	admin, err := this.adminRepo.FindOneByEmail(email)
	if err != nil || admin == nil {
		return nil, err
	}
	return &recoveryAccount{
		actor: common.Actor{ID: admin.ID, Email: admin.Email, FirstName: admin.FirstName,
			LastName: admin.LastName, PhoneNumber: admin.PhoneNumber, Role: admin.Role},
		passwordHash: admin.PasswordHash,
		savePassword: func(hash string) error { return this.adminRepo.UpdatePasswordHash(admin.ID, hash) },
	}, nil
}

func (this *service) findRecoveryAccountByActor(actor common.Actor) (*recoveryAccount, error) {
	if actor.Role == common.ActorRoleTenant {
		tenant, err := this.tenantRepo.FindOneById(actor.ID)
		if err != nil || tenant == nil {
			return nil, err
		}
		return &recoveryAccount{
			actor: common.Actor{ID: tenant.ID, Email: tenant.Email, FirstName: tenant.FirstName,
				LastName: tenant.LastName, PhoneNumber: tenant.PhoneNumber, Role: common.ActorRoleTenant},
			passwordHash: tenant.PasswordHash,
			savePassword: func(hash string) error { return this.tenantRepo.UpdatePasswordHash(tenant.ID, hash) },
		}, nil
	}
	admin, err := this.adminRepo.FindOneById(actor.ID)
	if err != nil || admin == nil {
		return nil, err
	}
	return &recoveryAccount{
		actor: common.Actor{ID: admin.ID, Email: admin.Email, FirstName: admin.FirstName,
			LastName: admin.LastName, PhoneNumber: admin.PhoneNumber, Role: admin.Role},
		passwordHash: admin.PasswordHash,
		savePassword: func(hash string) error { return this.adminRepo.UpdatePasswordHash(admin.ID, hash) },
	}, nil
}

// forgotPassword emails a reset link if the account exists. It answers the same
// way either way, so it can't be used to find out which emails have accounts.
func (this *service) forgotPassword(dto forgotPasswordReq) common.ServiceResponse {
	logger := this.logger.With("caller", "auth.service.forgotPassword")
	email := strings.TrimSpace(dto.Email)
	account, err := this.findRecoveryAccountByEmail(email, dto.AccountType)
	if err != nil {
		logger.Error("failed to look up account", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	if account == nil {
		logger.Info("password reset requested for an unknown account", "account_type", dto.AccountType)
		return common.ServiceResponse{StatusCode: 200}
	}

	token, err := common.GenerateResetJWT(*this.cfg, account.actor, passwordVersion(account.passwordHash))
	if err != nil {
		logger.Error("failed to generate reset token", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	link := this.cfg.WebAppURL + "/reset-password?token=" + token
	_, err = cpaas.QueueEmail(this.jobRepo, this.db, cpaas.EmailJob{
		Payload: map[string]string{"firstName": account.actor.FirstName, "link": link},
		Recipients: []cpaas.EmailRecipient{{
			Address: account.actor.Email,
			Name:    strings.TrimSpace(account.actor.FirstName + " " + account.actor.LastName),
		}},
		Subject: "Reset your Aventra password",
		Type:    cpaas.EmailTypePasswordReset,
	})
	if err != nil {
		logger.Error("failed to queue reset email", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	if this.cfg.Env == "development" {
		logger.Debug("password reset link", "actor", account.actor.Email, "link", link)
	}
	return common.ServiceResponse{StatusCode: 200}
}

// resetPassword sets a new password from a reset link and signs the user in.
func (this *service) resetPassword(dto resetPasswordReq) common.ServiceResponse {
	logger := this.logger.With("caller", "auth.service.resetPassword")
	claims, err := common.ParseJWT(dto.Token, []byte(this.cfg.JWTSecretKey), false)
	if err != nil || claims.TokenType != common.TokenTypeReset {
		return common.ServiceResponse{StatusCode: 401, Err: common.ErrInvalidToken}
	}
	account, err := this.findRecoveryAccountByActor(claims.Actor)
	if err != nil {
		logger.Error("failed to look up account", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	// Gone, or the password changed since the link was sent (including by
	// using this link already).
	if account == nil || passwordVersion(account.passwordHash) != claims.PasswordVersion {
		return common.ServiceResponse{StatusCode: 401, Err: common.ErrInvalidToken}
	}
	logger = logger.With("actor", account.actor.Email)

	hash, err := common.HashPassword(dto.Password)
	if err != nil {
		logger.Error("failed to hash password", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	if err := account.savePassword(hash); err != nil {
		logger.Error("failed to save password hash", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	accessToken, err := common.GenerateJWT(*this.cfg, account.actor)
	if err != nil {
		logger.Error("failed to generate access token", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	refreshToken, err := this.generateRefreshToken(account.actor)
	if err != nil {
		logger.Error("failed to generate refresh token", "err", err.Error())
		return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
	}
	return common.ServiceResponse{StatusCode: 200, Data: map[string]any{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	}}
}
