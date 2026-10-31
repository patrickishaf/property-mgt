package tenant

import (
	"testing"

	"github.com/patrickishaf/property-mgt/internal/db/model"
	"gorm.io/gorm"
)

func TestToTenantRes(t *testing.T) {
	tenant := model.Tenant{
		Model:                 gorm.Model{ID: 7},
		Email:                 "t@example.com",
		ExpectedAnnualRent:    1200,
		FirstName:             "Ada",
		LastName:              "Obi",
		LeaseID:               3,
		OutstandingAnnualRent: 400,
		PasswordHash:          "secret-hash",
		PhoneNumber:           "080",
		PropertyID:            5,
		UnitNumber:            "2B",
	}
	res := toTenantRes(tenant)
	want := tenantRes{
		ID:              7,
		Email:           "t@example.com",
		ExpectedRent:    1200,
		FirstName:       "Ada",
		LastName:        "Obi",
		LeaseID:         3,
		OutstandingRent: 400,
		PhoneNumber:     "080",
		PropertyID:      5,
		UnitNumber:      "2B",
	}
	if res != want {
		t.Fatalf("toTenantRes() = %+v, want %+v", res, want)
	}
}
