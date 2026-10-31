package property

import (
	"testing"
	"time"

	"github.com/patrickishaf/property-mgt/internal/db/model"
	"github.com/patrickishaf/property-mgt/internal/db/repository"
	"gorm.io/gorm"
)

func TestToPropertyRes(t *testing.T) {
	archived := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	p := model.Property{
		Model:                 gorm.Model{ID: 4},
		ArchivedAt:            &archived,
		ManagerID:             9,
		Name:                  "Palm Court",
		NumberOfOccupiedUnits: 1,
		NumberOfUnits:         3,
		OwnerID:               2,
	}
	res := toPropertyRes(p, repository.RentTotals{Collected: 300, Expected: 1200, Outstanding: 900})
	if res.ID != 4 || res.OwnerID != 2 || res.Name != "Palm Court" || res.ArchivedAt != &archived {
		t.Fatalf("unexpected identity fields: %+v", res)
	}
	if res.ManagerID == nil || *res.ManagerID != 9 {
		t.Fatalf("expected managerId 9, got %v", res.ManagerID)
	}
	if res.RentCollected != 300 || res.RentExpected != 1200 || res.OutstandingRent != 900 {
		t.Fatalf("unexpected rent totals: %+v", res)
	}
	if res.NumberOfUnits != 3 || res.NumberOfOccupiedUnits != 1 {
		t.Fatalf("unexpected unit counts: %+v", res)
	}

	p.ManagerID = 0
	if res := toPropertyRes(p, repository.RentTotals{}); res.ManagerID != nil {
		t.Fatalf("expected nil managerId for unmanaged property, got %v", *res.ManagerID)
	}
}

func TestToManagerResStatus(t *testing.T) {
	if got := toManagerRes(model.Admin{}).Status; got != managerStatusInvited {
		t.Fatalf("expected %q for manager without password, got %q", managerStatusInvited, got)
	}
	if got := toManagerRes(model.Admin{PasswordHash: "hash"}).Status; got != managerStatusActive {
		t.Fatalf("expected %q for manager with password, got %q", managerStatusActive, got)
	}
}
