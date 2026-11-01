package lease

import (
	"testing"
	"time"

	"github.com/patrickishaf/property-mgt/internal/db/model"
	"gorm.io/gorm"
)

func TestBuildDueItems(t *testing.T) {
	now := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	day := func(m time.Month, d int) *time.Time {
		v := time.Date(2026, m, d, 0, 0, 0, 0, time.UTC)
		return &v
	}
	installments := []model.Installment{
		{Model: gorm.Model{ID: 10}, LeaseID: 2, DueDate: day(12, 1), OutstandingAmount: 400},
		{Model: gorm.Model{ID: 11}, LeaseID: 2, DueDate: day(10, 15), OutstandingAmount: 150},
	}
	leases := []model.Lease{
		{Model: gorm.Model{ID: 1}, StartDate: *day(10, 1), OutstandingRent: 1200, UnitNumber: "1A"},
		{Model: gorm.Model{ID: 3}, StartDate: *day(12, 15), OutstandingRent: 900, UnitNumber: "2B"},
	}

	all := buildDueItems(installments, leases, now, "")
	got := ""
	for _, item := range all {
		got += item.DueDate + "/" + item.Type + " "
	}
	want := "2026-10-01/lease 2026-10-15/installment 2026-12-01/installment 2026-12-15/lease "
	if got != want {
		t.Fatalf("order = %q, want %q", got, want)
	}
	if !all[0].Overdue || !all[1].Overdue || all[2].Overdue || all[3].Overdue {
		t.Fatalf("unexpected overdue flags: %+v", all)
	}
	if all[0].UnitNumber != "1A" || all[0].InstallmentID != nil || all[1].InstallmentID == nil || *all[1].InstallmentID != 11 {
		t.Fatalf("unexpected item details: %+v", all[:2])
	}
	if all[1].Amount != 150 {
		t.Fatalf("installment amount should be what's still owed, got %v", all[1].Amount)
	}

	if overdue := buildDueItems(installments, leases, now, "overdue"); len(overdue) != 2 {
		t.Fatalf("expected 2 overdue items, got %d", len(overdue))
	}
	if upcoming := buildDueItems(installments, leases, now, "upcoming"); len(upcoming) != 2 || upcoming[0].Overdue {
		t.Fatalf("expected 2 upcoming items, got %+v", upcoming)
	}
}

func TestPaginate(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	if got := paginate(items, 2, 2); len(got) != 2 || got[0] != 3 {
		t.Fatalf("page 2 = %v", got)
	}
	if got := paginate(items, 3, 2); len(got) != 1 || got[0] != 5 {
		t.Fatalf("page 3 = %v", got)
	}
	if got := paginate(items, 4, 2); len(got) != 0 || got == nil {
		t.Fatalf("page past the end should be empty, not nil: %v", got)
	}
}
