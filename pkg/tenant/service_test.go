package tenant

import (
	"testing"

	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/internal/db/model"
	"gorm.io/gorm"
)

func TestCanAccessProperty(t *testing.T) {
	property := &model.Property{OwnerID: 1, ManagerID: 2}
	unmanaged := &model.Property{OwnerID: 1}

	cases := []struct {
		name     string
		actor    *common.Actor
		property *model.Property
		want     bool
	}{
		{"owner of property", &common.Actor{ID: 1, Role: common.AdminRoleOwner.ToString()}, property, true},
		{"other owner", &common.Actor{ID: 3, Role: common.AdminRoleOwner.ToString()}, property, false},
		{"manager of property", &common.Actor{ID: 2, Role: common.AdminRoleManager.ToString()}, property, true},
		{"manager matching owner id", &common.Actor{ID: 1, Role: common.AdminRoleManager.ToString()}, property, false},
		{"manager of unmanaged property", &common.Actor{ID: 0, Role: common.AdminRoleManager.ToString()}, unmanaged, false},
		{"tenant", &common.Actor{ID: 1, Role: common.ActorRoleTenant}, property, false},
		{"missing property", &common.Actor{ID: 1, Role: common.AdminRoleOwner.ToString()}, nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := canAccessProperty(tc.actor, tc.property); got != tc.want {
				t.Fatalf("canAccessProperty() = %v, want %v", got, tc.want)
			}
		})
	}
}

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
