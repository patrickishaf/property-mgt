package payment

import (
	"testing"
	"time"

	"github.com/patrickishaf/property-mgt/internal/db/model"
	"gorm.io/gorm"
)

func schedule(amounts ...float64) []model.Installment {
	installments := make([]model.Installment, 0, len(amounts))
	for i, amount := range amounts {
		installments = append(installments, model.Installment{
			Model:             gorm.Model{ID: uint(i + 1)},
			AmountExpected:    amount,
			OutstandingAmount: amount,
			Status:            model.InstallmentStatusUnpaid,
		})
	}
	return installments
}

func TestAllocateWithoutInstallments(t *testing.T) {
	res, err := allocate(allocationInput{Amount: 500, LeaseOutstanding: 1200, Purpose: model.PaymentPurposePartial})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Lines) != 1 || res.Lines[0].InstallmentID != nil || res.Lines[0].Amount != 500 {
		t.Fatalf("expected one lease-level line of 500, got %+v", res.Lines)
	}
	if res.LeaseOutstandingAfter != 700 || res.Status != model.PaymentStatusPartiallyPaid {
		t.Fatalf("unexpected result: %+v", res)
	}

	res, err = allocate(allocationInput{Amount: 1200, LeaseOutstanding: 1200, Purpose: model.PaymentPurposeLease})
	if err != nil || res.Status != model.PaymentStatusPaid || res.LeaseOutstandingAfter != 0 {
		t.Fatalf("expected full lease payment to be paid, got %+v, %v", res, err)
	}
}

func TestAllocateOldestFirst(t *testing.T) {
	res, err := allocate(allocationInput{
		Amount:           500,
		Installments:     schedule(400, 400, 400),
		LeaseOutstanding: 1200,
		Purpose:          model.PaymentPurposePartial,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Lines) != 2 || *res.Lines[0].InstallmentID != 1 || res.Lines[0].Amount != 400 ||
		*res.Lines[1].InstallmentID != 2 || res.Lines[1].Amount != 100 {
		t.Fatalf("expected 400 to #1 and 100 to #2, got %+v", res.Lines)
	}
	if res.Installments[0].Status != model.InstallmentStatusPaid || res.Installments[1].Status != model.InstallmentStatusPartiallyPaid {
		t.Fatalf("unexpected installment statuses: %+v", res.Installments)
	}
	if res.Installments[1].OutstandingAmount != 300 || res.Installments[1].LastPaymentDate == nil {
		t.Fatalf("unexpected second installment: %+v", res.Installments[1])
	}
	if res.Status != model.PaymentStatusPartiallyPaid || res.LeaseOutstandingAfter != 700 {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestAllocateSpecificInstallments(t *testing.T) {
	res, err := allocate(allocationInput{
		Amount:           400,
		InstallmentIDs:   []uint{3},
		Installments:     schedule(400, 400, 400),
		LeaseOutstanding: 1200,
		Purpose:          model.PaymentPurposeInstallments,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Lines) != 1 || *res.Lines[0].InstallmentID != 3 {
		t.Fatalf("expected payment to go to installment #3 only, got %+v", res.Lines)
	}
	if res.Status != model.PaymentStatusPaid {
		t.Fatalf("expected paid, got %s", res.Status)
	}
}

func TestAllocateTargetedShortfallIsPartial(t *testing.T) {
	res, err := allocate(allocationInput{
		Amount:           300,
		InstallmentIDs:   []uint{2},
		Installments:     schedule(400, 400),
		LeaseOutstanding: 800,
		Purpose:          model.PaymentPurposeInstallments,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != model.PaymentStatusPartiallyPaid || res.Installments[0].ID != 2 {
		t.Fatalf("expected partial payment on #2, got %+v", res)
	}
}

func TestAllocateSpillsOverToOtherInstallments(t *testing.T) {
	installments := schedule(400, 400)
	installments[1].AmountPaid = 300
	installments[1].OutstandingAmount = 100
	res, err := allocate(allocationInput{
		Amount:           400,
		InstallmentIDs:   []uint{2},
		Installments:     installments,
		LeaseOutstanding: 500,
		Purpose:          model.PaymentPurposeInstallments,
		Now:              time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Lines) != 2 || *res.Lines[0].InstallmentID != 2 || res.Lines[0].Amount != 100 ||
		*res.Lines[1].InstallmentID != 1 || res.Lines[1].Amount != 300 {
		t.Fatalf("expected 100 to #2 then 300 to #1, got %+v", res.Lines)
	}
	if res.Status != model.PaymentStatusPaid {
		t.Fatalf("expected paid since #2 is settled, got %s", res.Status)
	}
}

func TestAllocateRejectsOverpayment(t *testing.T) {
	if _, err := allocate(allocationInput{Amount: 1300, LeaseOutstanding: 1200}); err != errAmountExceedsOutstanding {
		t.Fatalf("expected errAmountExceedsOutstanding, got %v", err)
	}
	if _, err := allocate(allocationInput{Amount: 0, LeaseOutstanding: 1200}); err != errAmountExceedsOutstanding {
		t.Fatalf("expected zero amount to be rejected, got %v", err)
	}
}

func TestAllocateRounding(t *testing.T) {
	res, err := allocate(allocationInput{
		Amount:           0.3,
		Installments:     schedule(0.1, 0.2),
		LeaseOutstanding: 0.30000000000000004,
		Purpose:          model.PaymentPurposeLease,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.LeaseOutstandingAfter != 0 || res.Status != model.PaymentStatusPaid {
		t.Fatalf("expected rounding to settle the lease, got %+v", res)
	}
}
