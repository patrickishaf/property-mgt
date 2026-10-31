package payment

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/config"
	"github.com/patrickishaf/property-mgt/internal/db/model"
	"github.com/patrickishaf/property-mgt/internal/db/repository"
	"github.com/patrickishaf/property-mgt/providers/bachs"
	"gorm.io/gorm"
)

const (
	errInvalidInstallment = "invalid installment"
	errLeaseNotActive     = "lease is not active"
	errLeaseNotFound      = "lease not found"
	errNotFound           = "payment not found"
)

type service struct {
	bachsClient     *bachs.Client
	cfg             *config.AppConfig
	db              *gorm.DB
	installmentRepo *repository.InstallmentRepository
	leaseRepo       *repository.LeaseRepository
	logger          *slog.Logger
	paymentRepo     *repository.PaymentRepository
	propertyRepo    *repository.PropertyRepository
	tenantRepo      *repository.TenantRepository
}

func CreateService(
	cfg *config.AppConfig,
	logger *slog.Logger,
	db *gorm.DB,
	paymentRepo *repository.PaymentRepository,
	installmentRepo *repository.InstallmentRepository,
	leaseRepo *repository.LeaseRepository,
	propertyRepo *repository.PropertyRepository,
	tenantRepo *repository.TenantRepository,
	bachsClient *bachs.Client,
) *service {
	return &service{
		bachsClient:     bachsClient,
		cfg:             cfg,
		db:              db,
		installmentRepo: installmentRepo,
		leaseRepo:       leaseRepo,
		logger:          logger,
		paymentRepo:     paymentRepo,
		propertyRepo:    propertyRepo,
		tenantRepo:      tenantRepo,
	}
}

func isAdmin(actor *common.Actor) bool {
	return actor.Role == common.AdminRoleOwner.ToString() || actor.Role == common.AdminRoleManager.ToString()
}

func forbidden() common.ServiceResponse {
	return common.ServiceResponse{StatusCode: 403, Err: common.ErrInvalidCredentials}
}

func serverError() common.ServiceResponse {
	return common.ServiceResponse{StatusCode: 500, Err: common.ErrGeneric}
}

func newReference() string {
	return "PAY-" + strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", "")[:16])
}

// adminScope returns the owner/manager IDs that limit queries to the actor's properties.
func adminScope(actor *common.Actor) (ownerID uint, managerID uint) {
	if actor.Role == common.AdminRoleOwner.ToString() {
		return actor.ID, 0
	}
	return 0, actor.ID
}

// findAdminLease loads a lease on a property the admin owns or manages.
func (this *service) findAdminLease(actor *common.Actor, leaseID uint, logger *slog.Logger) (*model.Lease, *model.Property, *common.ServiceResponse) {
	if !isAdmin(actor) {
		res := forbidden()
		return nil, nil, &res
	}
	lease, err := this.leaseRepo.FindOneById(leaseID)
	if err != nil {
		logger.Error("failed to find lease", "err", err.Error())
		res := serverError()
		return nil, nil, &res
	}
	if lease == nil {
		return nil, nil, &common.ServiceResponse{StatusCode: 404, Err: errLeaseNotFound}
	}
	property, err := this.propertyRepo.FindOneById(uint(lease.PropertyID))
	if err != nil {
		logger.Error("failed to find property", "err", err.Error())
		res := serverError()
		return nil, nil, &res
	}
	if !property.AccessibleBy(actor) {
		return nil, nil, &common.ServiceResponse{StatusCode: 404, Err: errLeaseNotFound}
	}
	return lease, property, nil
}

// canSeeRecord reports whether the actor may read a payment or installment
// belonging to the given tenant and property.
func (this *service) canSeeRecord(actor *common.Actor, tenantID int, propertyID int) (bool, error) {
	if actor.Role == common.ActorRoleTenant {
		return uint(tenantID) == actor.ID, nil
	}
	if !isAdmin(actor) {
		return false, nil
	}
	property, err := this.propertyRepo.FindOneById(uint(propertyID))
	if err != nil {
		return false, err
	}
	return property.AccessibleBy(actor), nil
}

// validateInstallments checks that every ID is an installment of the lease that
// still has money owed, and returns their total outstanding amount.
func validateInstallments(ids []uint, installments []model.Installment) (float64, bool) {
	byID := make(map[uint]model.Installment, len(installments))
	for _, inst := range installments {
		byID[inst.ID] = inst
	}
	seen := make(map[uint]bool, len(ids))
	total := 0.0
	for _, id := range ids {
		inst, ok := byID[id]
		if !ok || seen[id] || inst.OutstandingAmount <= 0 {
			return 0, false
		}
		seen[id] = true
		total += inst.OutstandingAmount
	}
	return round2(total), true
}

func joinIDs(ids []uint) string {
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, strconv.FormatUint(uint64(id), 10))
	}
	return strings.Join(parts, ",")
}

func splitIDs(value string) []uint {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	ids := make([]uint, 0, len(parts))
	for _, p := range parts {
		if id, err := strconv.ParseUint(p, 10, 64); err == nil {
			ids = append(ids, uint(id))
		}
	}
	return ids
}

func formatAmount(amount float64) string {
	return strconv.FormatFloat(round2(amount), 'f', 2, 64)
}

// applyPayment allocates a saved payment to its lease inside the transaction:
// it locks the lease and its installments, updates their balances and the
// tenant's outstanding rent, records allocations and a receipt, and sets the
// payment's final status. errAmountExceedsOutstanding means the payment no
// longer fits the lease's balance.
func (this *service) applyPayment(trx *gorm.DB, payment *model.Payment, paidAt time.Time) error {
	lease, err := this.leaseRepo.FindOneByIdForUpdateTRX(uint(payment.LeaseID), trx)
	if err != nil {
		return err
	}
	if lease == nil {
		return errors.New(errLeaseNotFound)
	}
	installments, err := this.installmentRepo.FindByLeaseForUpdateTRX(lease.ID, trx)
	if err != nil {
		return err
	}
	result, err := allocate(allocationInput{
		Amount:           payment.Amount,
		InstallmentIDs:   splitIDs(payment.InstallmentIDs),
		Installments:     installments,
		LeaseOutstanding: lease.OutstandingRent,
		Now:              paidAt,
		Purpose:          payment.Purpose,
	})
	if err != nil {
		return err
	}

	for _, inst := range result.Installments {
		err := this.installmentRepo.UpdateTRX(inst.ID, map[string]any{
			"amount_paid":        inst.AmountPaid,
			"outstanding_amount": inst.OutstandingAmount,
			"status":             inst.Status,
			"last_payment_date":  inst.LastPaymentDate,
		}, trx)
		if err != nil {
			return err
		}
	}
	previouslyPaid := lease.AmountPaid
	err = this.leaseRepo.UpdateTRX(lease.ID, map[string]any{
		"amount_paid":      round2(lease.AmountPaid + payment.Amount),
		"outstanding_rent": result.LeaseOutstandingAfter,
	}, trx)
	if err != nil {
		return err
	}
	err = this.tenantRepo.UpdateTRX(uint(lease.TenantID), map[string]any{
		"outstanding_annual_rent": result.LeaseOutstandingAfter,
	}, trx)
	if err != nil {
		return err
	}

	allocations := make([]model.PaymentAllocation, 0, len(result.Lines))
	for _, line := range result.Lines {
		allocations = append(allocations, model.PaymentAllocation{
			Amount:        line.Amount,
			InstallmentID: line.InstallmentID,
			PaymentID:     payment.ID,
		})
	}
	if err := this.paymentRepo.SaveAllocationsTRX(allocations, trx); err != nil {
		return err
	}

	payment.Status = result.Status
	payment.PaidAt = &paidAt
	err = this.paymentRepo.UpdateTRX(payment.ID, map[string]any{
		"status":  payment.Status,
		"paid_at": payment.PaidAt,
	}, trx)
	if err != nil {
		return err
	}

	receipt := model.Receipt{
		AmountPaid:                    formatAmount(payment.Amount),
		OutstandingInstallmentPayment: formatAmount(result.LeaseOutstandingAfter),
		PaymentDate:                   paidAt.Format(time.RFC3339),
		PaymentID:                     payment.ID,
		PaymentMethod:                 payment.Method,
		PreviousPaymentAmount:         formatAmount(previouslyPaid),
		SenderFullName:                payment.TenantFullName,
	}
	if tenant, err := this.tenantRepo.FindOneById(uint(lease.TenantID)); err == nil && tenant != nil {
		receipt.SenderEmail = tenant.Email
	}
	if n := len(result.Installments); n > 0 {
		// The receipt describes the last installment the payment reached.
		last := result.Installments[n-1]
		receipt.InstallmentAmount = formatAmount(last.AmountExpected)
		receipt.InstallmentAmountPaid = formatAmount(last.AmountPaid)
		receipt.InstallmentStatus = last.Status
		receipt.OutstandingInstallmentPayment = formatAmount(last.OutstandingAmount)
		if last.DueDate != nil {
			receipt.InstallmentDueDate = last.DueDate.Format(dateLayout)
		}
	}
	return this.paymentRepo.SaveReceiptTRX(&receipt, trx)
}

func (this *service) createPayment(actor *common.Actor, dto createPaymentReq) common.ServiceResponse {
	logger := this.logger.With("caller", "payment.service.createPayment", "actor", actor.Email)
	lease, property, errRes := this.findAdminLease(actor, dto.LeaseID, logger)
	if errRes != nil {
		return *errRes
	}
	if lease.Status != model.LeaseStatusActive {
		return common.ServiceResponse{StatusCode: 409, Err: errLeaseNotActive}
	}
	reference := dto.Reference
	if reference == "" {
		reference = newReference()
	} else if exists, err := this.paymentRepo.ExistsByReference(reference); err != nil {
		logger.Error("failed to check reference", "err", err.Error())
		return serverError()
	} else if exists {
		return common.ServiceResponse{StatusCode: 409, Err: "reference already exists"}
	}

	amount := round2(dto.Amount)
	if amount > round2(lease.OutstandingRent) {
		return common.ServiceResponse{StatusCode: 400, Err: errAmountExceedsOutstanding.Error()}
	}

	purpose := model.PaymentPurposeLease
	target := round2(lease.OutstandingRent)
	if len(dto.InstallmentIDs) > 0 {
		installments, err := this.installmentRepo.FindByLease(lease.ID)
		if err != nil {
			logger.Error("failed to find installments", "err", err.Error())
			return serverError()
		}
		total, ok := validateInstallments(dto.InstallmentIDs, installments)
		if !ok {
			return common.ServiceResponse{StatusCode: 400, Err: errInvalidInstallment}
		}
		purpose = model.PaymentPurposeInstallments
		target = total
	}

	paidAt := time.Now()
	if dto.PaidAt != "" {
		paidAt, _ = time.Parse(dateLayout, dto.PaidAt)
	}
	tenant, err := this.tenantRepo.FindOneById(uint(lease.TenantID))
	if err != nil || tenant == nil {
		logger.Error("failed to find tenant for lease", "lease_id", lease.ID)
		return serverError()
	}

	recordedBy := actor.ID
	payment := model.Payment{
		Amount:         amount,
		Currency:       this.cfg.PaymentCurrency,
		InstallmentIDs: joinIDs(dto.InstallmentIDs),
		LeaseID:        int(lease.ID),
		Method:         dto.Method,
		Note:           dto.Note,
		PropertyID:     lease.PropertyID,
		PropertyName:   property.Name,
		Purpose:        purpose,
		RecordedBy:     &recordedBy,
		Reference:      reference,
		Source:         model.PaymentSourceOffline,
		Status:         model.PaymentStatusUnpaid,
		TargetAmount:   target,
		TenantFullName: tenant.FirstName + " " + tenant.LastName,
		TenantID:       lease.TenantID,
		UnitNumber:     lease.UnitNumber,
	}
	err = this.db.Transaction(func(trx *gorm.DB) error {
		if err := this.paymentRepo.SaveOneTRX(&payment, trx); err != nil {
			return err
		}
		return this.applyPayment(trx, &payment, paidAt)
	})
	if errors.Is(err, errAmountExceedsOutstanding) {
		return common.ServiceResponse{StatusCode: 400, Err: errAmountExceedsOutstanding.Error()}
	}
	if err != nil {
		logger.Error("failed to record payment", "err", err.Error())
		return serverError()
	}
	return this.paymentDetail(&payment, 201, logger)
}

func (this *service) paymentDetail(payment *model.Payment, statusCode int, logger *slog.Logger) common.ServiceResponse {
	allocations, err := this.paymentRepo.FindAllocations(payment.ID)
	if err != nil {
		logger.Error("failed to find allocations", "err", err.Error())
		return serverError()
	}
	receipt, err := this.paymentRepo.FindReceipt(payment.ID)
	if err != nil {
		logger.Error("failed to find receipt", "err", err.Error())
		return serverError()
	}
	res := paymentDetailRes{paymentRes: toPaymentRes(*payment), Allocations: make([]allocationRes, 0, len(allocations))}
	for _, a := range allocations {
		res.Allocations = append(res.Allocations, allocationRes{Amount: a.Amount, InstallmentID: a.InstallmentID})
	}
	if receipt != nil {
		r := toReceiptRes(*receipt)
		res.Receipt = &r
	}
	return common.ServiceResponse{StatusCode: statusCode, Data: res}
}

func parseDate(value string) *time.Time {
	if value == "" {
		return nil
	}
	t, err := time.Parse(dateLayout, value)
	if err != nil {
		return nil
	}
	return &t
}

func (this *service) getPayments(actor *common.Actor, dto getPaymentsReq) common.ServiceResponse {
	logger := this.logger.With("caller", "payment.service.getPayments", "actor", actor.Email)
	filter := repository.PaymentFilter{
		From:       parseDate(dto.From),
		LeaseID:    dto.LeaseID,
		Method:     dto.Method,
		PropertyID: dto.PropertyID,
		Purpose:    dto.Purpose,
		Source:     dto.Source,
		Status:     dto.Status,
		TenantID:   dto.TenantID,
	}
	if to := parseDate(dto.To); to != nil {
		end := to.AddDate(0, 0, 1)
		filter.To = &end
	}
	switch {
	case actor.Role == common.ActorRoleTenant:
		filter.TenantID = actor.ID
	case isAdmin(actor):
		filter.OwnerID, filter.ManagerID = adminScope(actor)
	default:
		return forbidden()
	}

	if err := this.paymentRepo.ExpireStaleCheckouts(time.Now()); err != nil {
		logger.Error("failed to expire stale checkouts", "err", err.Error())
	}
	page, limit := dto.values()
	payments, total, err := this.paymentRepo.FindMany(filter, page, limit)
	if err != nil {
		logger.Error("failed to find payments", "err", err.Error())
		return serverError()
	}
	items := make([]paymentRes, 0, len(payments))
	for _, p := range payments {
		items = append(items, toPaymentRes(p))
	}
	return common.ServiceResponse{StatusCode: 200, Data: listRes[paymentRes]{Items: items, Limit: limit, Page: page, Total: total}}
}

func (this *service) getPaymentInfo(actor *common.Actor, id uint) common.ServiceResponse {
	logger := this.logger.With("caller", "payment.service.getPaymentInfo", "actor", actor.Email)
	if err := this.paymentRepo.ExpireStaleCheckouts(time.Now()); err != nil {
		logger.Error("failed to expire stale checkouts", "err", err.Error())
	}
	payment, err := this.paymentRepo.FindOneById(id)
	if err != nil {
		logger.Error("failed to find payment", "err", err.Error())
		return serverError()
	}
	if payment == nil {
		return common.ServiceResponse{StatusCode: 404, Err: errNotFound}
	}
	ok, err := this.canSeeRecord(actor, payment.TenantID, payment.PropertyID)
	if err != nil {
		logger.Error("failed to check payment access", "err", err.Error())
		return serverError()
	}
	if !ok {
		return common.ServiceResponse{StatusCode: 404, Err: errNotFound}
	}
	return this.paymentDetail(payment, 200, logger)
}

func (this *service) getTotalPaymentStats(actor *common.Actor, dto getStatsReq) common.ServiceResponse {
	logger := this.logger.With("caller", "payment.service.getTotalPaymentStats", "actor", actor.Email)
	if !isAdmin(actor) {
		return forbidden()
	}
	ownerID, managerID := adminScope(actor)
	if dto.PropertyID != 0 {
		property, err := this.propertyRepo.FindOneById(dto.PropertyID)
		if err != nil {
			logger.Error("failed to find property", "err", err.Error())
			return serverError()
		}
		if !property.AccessibleBy(actor) {
			return common.ServiceResponse{StatusCode: 404, Err: "property not found"}
		}
	}

	totals, err := this.leaseRepo.PortfolioTotals(ownerID, managerID, dto.PropertyID)
	if err != nil {
		logger.Error("failed to sum lease totals", "err", err.Error())
		return serverError()
	}
	instStats, err := this.installmentRepo.Stats(ownerID, managerID, dto.PropertyID, time.Now())
	if err != nil {
		logger.Error("failed to compute installment stats", "err", err.Error())
		return serverError()
	}
	percentage := 0.0
	if totals.Expected > 0 {
		percentage = round2(totals.Collected / totals.Expected * 100)
	}
	return common.ServiceResponse{StatusCode: 200, Data: portfolioRes{
		ActiveTenantCount:       totals.ActiveTenantCount,
		OpenInstallmentCount:    instStats.OpenCount,
		OutstandingRentBalance:  round2(totals.Outstanding),
		OverdueInstallmentCount: instStats.OverdueCount,
		PercentageRentCollected: percentage,
		TotalOverdueAmount:      round2(instStats.OverdueAmount),
		TotalRentCollected:      round2(totals.Collected),
		TotalRentExpected:       round2(totals.Expected),
	}}
}

// describeOutstanding formats a balance for error messages.
func describeOutstanding(amount float64) string {
	return fmt.Sprintf("%.2f", round2(amount))
}
