package payment

import (
	"errors"
	"strconv"
	"time"

	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/internal/db/model"
	"github.com/patrickishaf/property-mgt/providers/bachs"
	"gorm.io/gorm"
)

func (this *service) initiateCheckout(actor *common.Actor, dto initiateCheckoutReq) common.ServiceResponse {
	logger := this.logger.With("caller", "payment.service.initiateCheckout", "actor", actor.Email)
	if actor.Role != common.ActorRoleTenant {
		return forbidden()
	}
	lease, err := this.leaseRepo.FindActiveByTenant(actor.ID)
	if err != nil {
		logger.Error("failed to find active lease", "err", err.Error())
		return serverError()
	}
	if lease == nil {
		return common.ServiceResponse{StatusCode: 404, Err: "no active lease"}
	}
	outstanding := round2(lease.OutstandingRent)
	if outstanding <= 0 {
		return common.ServiceResponse{StatusCode: 409, Err: "lease has no outstanding balance"}
	}

	var amount float64
	switch dto.Type {
	case model.PaymentPurposeLease:
		amount = outstanding
	case model.PaymentPurposeInstallments:
		if len(dto.InstallmentIDs) == 0 {
			return common.ServiceResponse{StatusCode: 400, Err: common.ErrInvalidRequest}
		}
		installments, err := this.installmentRepo.FindByLease(lease.ID)
		if err != nil {
			logger.Error("failed to find installments", "err", err.Error())
			return serverError()
		}
		total, ok := validateInstallments(dto.InstallmentIDs, installments)
		if !ok {
			return common.ServiceResponse{StatusCode: 400, Err: errInvalidInstallment}
		}
		amount = total
	case model.PaymentPurposePartial:
		if !lease.AllowPartialPayment {
			return common.ServiceResponse{StatusCode: 400, Err: "partial payments are not allowed for this lease"}
		}
		amount = round2(dto.Amount)
		if amount <= 0 {
			return common.ServiceResponse{StatusCode: 400, Err: common.ErrInvalidRequest}
		}
		if amount < round2(lease.MinimumPartialAmount) {
			return common.ServiceResponse{StatusCode: 400, Err: "amount is below the minimum partial amount of " + describeOutstanding(lease.MinimumPartialAmount)}
		}
		if amount > outstanding {
			return common.ServiceResponse{StatusCode: 400, Err: errAmountExceedsOutstanding.Error()}
		}
	}

	open, err := this.paymentRepo.FindOpenCheckoutForLease(lease.ID, time.Now())
	if err != nil {
		logger.Error("failed to find open checkout", "err", err.Error())
		return serverError()
	}
	if open != nil {
		return common.ServiceResponse{StatusCode: 409, Err: "a checkout is already in progress for this lease"}
	}
	property, err := this.propertyRepo.FindOneById(uint(lease.PropertyID))
	if err != nil || property == nil {
		logger.Error("failed to find property for lease", "lease_id", lease.ID)
		return serverError()
	}
	tenant, err := this.tenantRepo.FindOneById(actor.ID)
	if err != nil || tenant == nil {
		logger.Error("failed to find tenant", "tenant_id", actor.ID)
		return serverError()
	}

	installmentIDs := ""
	if dto.Type == model.PaymentPurposeInstallments {
		installmentIDs = joinIDs(dto.InstallmentIDs)
	}
	payment := model.Payment{
		Amount:         amount,
		Currency:       this.cfg.PaymentCurrency,
		InstallmentIDs: installmentIDs,
		LeaseID:        int(lease.ID),
		Method:         model.PaymentMethodCheckout,
		PropertyID:     lease.PropertyID,
		PropertyName:   property.Name,
		Purpose:        dto.Type,
		Reference:      newReference(),
		Source:         model.PaymentSourceCheckout,
		Status:         model.PaymentStatusUnpaid,
		TargetAmount:   amount,
		TenantFullName: tenant.FirstName + " " + tenant.LastName,
		TenantID:       int(tenant.ID),
		UnitNumber:     lease.UnitNumber,
	}
	if err := this.paymentRepo.SaveOneTRX(&payment, this.db); err != nil {
		logger.Error("failed to save checkout payment", "err", err.Error())
		return serverError()
	}

	body := bachs.CreateCheckoutSessionBody{
		BillingCurrency: this.cfg.PaymentCurrency,
		Metadata:        map[string]string{"paymentId": formatID(payment.ID), "leaseId": formatID(lease.ID)},
		Reference:       payment.Reference,
	}
	body.Pricing.Amount = formatAmount(amount)
	body.Pricing.Currency = this.cfg.PaymentCurrency
	body.Customer.Email = tenant.Email
	body.Customer.Name = payment.TenantFullName
	body.Customer.PhoneNumber = tenant.PhoneNumber
	checkout, err := this.bachsClient.InitiateCheckout(body)
	if err != nil {
		logger.Error("failed to initiate checkout", "err", err.Error())
		_ = this.paymentRepo.UpdateTRX(payment.ID, map[string]any{"status": model.PaymentStatusFailed}, this.db)
		return common.ServiceResponse{StatusCode: 502, Err: "payment provider unavailable"}
	}

	expiresAt, err := time.Parse(time.RFC3339, checkout.ExpiresAt)
	if err != nil {
		expiresAt = time.Now().Add(time.Duration(this.cfg.CheckoutExpiryMinutes) * time.Minute)
	}
	payment.CheckoutID = checkout.CheckoutID
	payment.CheckoutURL = checkout.CheckoutURL
	payment.ExpiresAt = &expiresAt
	err = this.paymentRepo.UpdateTRX(payment.ID, map[string]any{
		"checkout_id":  payment.CheckoutID,
		"checkout_url": payment.CheckoutURL,
		"expires_at":   payment.ExpiresAt,
	}, this.db)
	if err != nil {
		logger.Error("failed to save checkout details", "err", err.Error())
		return serverError()
	}
	return common.ServiceResponse{StatusCode: 201, Data: toPaymentRes(payment)}
}

var (
	errCheckoutExpired     = errors.New("checkout has expired")
	errCheckoutNotUnpaid   = errors.New("payment is not awaiting completion")
	errCheckoutNotFound    = errors.New(errNotFound)
	errCheckoutNotCheckout = errors.New("payment is not a checkout")
)

// completeCheckout settles an unpaid checkout as if the provider had confirmed
// it. A Bachs webhook should call this once the integration exists.
func (this *service) completeCheckout(reference string) common.ServiceResponse {
	logger := this.logger.With("caller", "payment.service.completeCheckout", "reference", reference)
	now := time.Now()
	var payment *model.Payment
	err := this.db.Transaction(func(trx *gorm.DB) error {
		var err error
		payment, err = this.paymentRepo.FindOneByReferenceForUpdateTRX(reference, trx)
		if err != nil {
			return err
		}
		if payment == nil {
			return errCheckoutNotFound
		}
		if payment.Source != model.PaymentSourceCheckout {
			return errCheckoutNotCheckout
		}
		if payment.Status != model.PaymentStatusUnpaid {
			return errCheckoutNotUnpaid
		}
		if payment.ExpiresAt != nil && !payment.ExpiresAt.After(now) {
			return errCheckoutExpired
		}
		return this.applyPayment(trx, payment, now)
	})
	switch {
	case err == nil:
		return this.paymentDetail(payment, 200, logger)
	case errors.Is(err, errCheckoutNotFound):
		return common.ServiceResponse{StatusCode: 404, Err: errNotFound}
	case errors.Is(err, errCheckoutExpired):
		if uerr := this.paymentRepo.UpdateTRX(payment.ID, map[string]any{"status": model.PaymentStatusExpired}, this.db); uerr != nil {
			logger.Error("failed to mark checkout expired", "err", uerr.Error())
		}
		return common.ServiceResponse{StatusCode: 409, Err: err.Error()}
	case errors.Is(err, errCheckoutNotUnpaid), errors.Is(err, errCheckoutNotCheckout), errors.Is(err, errAmountExceedsOutstanding):
		return common.ServiceResponse{StatusCode: 409, Err: err.Error()}
	default:
		logger.Error("failed to complete checkout", "err", err.Error())
		return serverError()
	}
}

func formatID(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}
