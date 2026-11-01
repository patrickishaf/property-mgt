package payment

import (
	"strconv"

	"github.com/patrickishaf/property-mgt/internal/db/model"
)

// Every Bachs checkout carries the same metadata keys, and Bachs sends them back
// on its webhooks. Values are strings because Bachs metadata is a string map.
const (
	// metadataKind marks a checkout as an Aventra rent payment, as opposed to
	// anything else that may share the Bachs account later.
	metadataKind          = "rent_payment"
	metadataSchemaVersion = "1"

	metadataTargetLease        = "lease"
	metadataTargetInstallments = "installments"
)

// paymentMetadata says what a checkout pays for. TargetType and TargetIDs name
// the thing being paid in the same two keys whatever the purpose: the lease for
// a full or partial lease payment, or the installments for an installment
// payment. LeaseID is always set, since installments belong to a lease.
type paymentMetadata struct {
	Kind          string // always metadataKind
	SchemaVersion string
	PaymentID     string // the payment record the checkout settles
	Purpose       string // lease | installments | partial
	TargetType    string // lease | installments
	TargetIDs     string // comma-separated IDs of TargetType, e.g. "12" or "31,32"
	LeaseID       string
	TenantID      string
	PropertyID    string
}

func checkoutMetadata(payment model.Payment, installmentIDs []uint) paymentMetadata {
	meta := paymentMetadata{
		Kind:          metadataKind,
		SchemaVersion: metadataSchemaVersion,
		PaymentID:     formatID(payment.ID),
		Purpose:       payment.Purpose,
		TargetType:    metadataTargetLease,
		TargetIDs:     strconv.Itoa(payment.LeaseID),
		LeaseID:       strconv.Itoa(payment.LeaseID),
		TenantID:      strconv.Itoa(payment.TenantID),
		PropertyID:    strconv.Itoa(payment.PropertyID),
	}
	if payment.Purpose == model.PaymentPurposeInstallments {
		meta.TargetType = metadataTargetInstallments
		meta.TargetIDs = joinIDs(installmentIDs)
	}
	return meta
}

func (this paymentMetadata) toMap() map[string]string {
	return map[string]string{
		"kind":           this.Kind,
		"schema_version": this.SchemaVersion,
		"payment_id":     this.PaymentID,
		"purpose":        this.Purpose,
		"target_type":    this.TargetType,
		"target_ids":     this.TargetIDs,
		"lease_id":       this.LeaseID,
		"tenant_id":      this.TenantID,
		"property_id":    this.PropertyID,
	}
}
