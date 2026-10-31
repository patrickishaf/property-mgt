package payment

import (
	"errors"
	"math"
	"time"

	"github.com/patrickishaf/property-mgt/internal/db/model"
)

var errAmountExceedsOutstanding = errors.New("amount exceeds outstanding balance")

// round2 rounds a money amount to 2 decimal places.
func round2(amount float64) float64 {
	return math.Round(amount*100) / 100
}

type allocationInput struct {
	Amount float64
	// InstallmentIDs are the installments a payment with purpose "installments" targets.
	InstallmentIDs []uint
	// Installments are all of the lease's installments, in due-date order.
	Installments     []model.Installment
	LeaseOutstanding float64
	Now              time.Time
	Purpose          string
}

type allocationLine struct {
	Amount        float64
	InstallmentID *uint
}

type allocationResult struct {
	Lines []allocationLine
	// Installments holds the updated copies of installments that received money.
	Installments          []model.Installment
	LeaseOutstandingAfter float64
	Status                string
}

// allocate works out how a payment is spread across a lease and its installments.
//
// Payments for specific installments go to those installments first. Lease and
// partial payments go to the oldest-due installments, or straight to the lease
// when it has no installments. Any amount left after the targeted installments
// spills over to the remaining installments, oldest first. The amount may not
// exceed the lease's outstanding rent.
func allocate(in allocationInput) (allocationResult, error) {
	amount := round2(in.Amount)
	outstanding := round2(in.LeaseOutstanding)
	if amount <= 0 || amount > outstanding {
		return allocationResult{}, errAmountExceedsOutstanding
	}

	res := allocationResult{LeaseOutstandingAfter: round2(outstanding - amount)}
	if len(in.Installments) == 0 {
		res.Lines = []allocationLine{{Amount: amount}}
		res.Status = model.PaymentStatusPartiallyPaid
		if res.LeaseOutstandingAfter == 0 {
			res.Status = model.PaymentStatusPaid
		}
		return res, nil
	}

	targeted := make(map[uint]bool, len(in.InstallmentIDs))
	for _, id := range in.InstallmentIDs {
		targeted[id] = true
	}
	installments := make([]model.Installment, len(in.Installments))
	copy(installments, in.Installments)

	remaining := amount
	touched := make(map[int]bool)
	apply := func(onlyTargeted bool) {
		for i := range installments {
			inst := &installments[i]
			if remaining <= 0 {
				return
			}
			if inst.OutstandingAmount <= 0 || (onlyTargeted && !targeted[inst.ID]) {
				continue
			}
			paid := math.Min(remaining, inst.OutstandingAmount)
			inst.AmountPaid = round2(inst.AmountPaid + paid)
			inst.OutstandingAmount = round2(inst.OutstandingAmount - paid)
			inst.Status = model.InstallmentStatusPartiallyPaid
			if inst.OutstandingAmount == 0 {
				inst.Status = model.InstallmentStatusPaid
			}
			now := in.Now
			inst.LastPaymentDate = &now
			remaining = round2(remaining - paid)
			id := inst.ID
			res.Lines = append(res.Lines, allocationLine{Amount: round2(paid), InstallmentID: &id})
			touched[i] = true
		}
	}
	if in.Purpose == model.PaymentPurposeInstallments {
		apply(true)
	}
	apply(false)
	if remaining > 0 {
		// Installments should always cover the lease's outstanding rent; if they
		// don't, refuse rather than lose track of money.
		return allocationResult{}, errAmountExceedsOutstanding
	}

	for i := range installments {
		if touched[i] {
			res.Installments = append(res.Installments, installments[i])
		}
	}

	res.Status = model.PaymentStatusPaid
	if in.Purpose == model.PaymentPurposeInstallments {
		for _, inst := range installments {
			if targeted[inst.ID] && inst.OutstandingAmount > 0 {
				res.Status = model.PaymentStatusPartiallyPaid
			}
		}
	} else if res.LeaseOutstandingAfter > 0 {
		res.Status = model.PaymentStatusPartiallyPaid
	}
	return res, nil
}
