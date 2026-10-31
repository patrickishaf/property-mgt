package model

import (
	"testing"

	"github.com/patrickishaf/property-mgt/common"
)

func TestPropertyAccess(t *testing.T) {
	property := &Property{OwnerID: 1, ManagerID: 2}
	unmanaged := &Property{OwnerID: 1}
	owner := &common.Actor{ID: 1, Role: common.AdminRoleOwner.ToString()}
	otherOwner := &common.Actor{ID: 3, Role: common.AdminRoleOwner.ToString()}
	manager := &common.Actor{ID: 2, Role: common.AdminRoleManager.ToString()}

	cases := []struct {
		name        string
		actor       *common.Actor
		property    *Property
		wantAccess  bool
		wantOwnedBy bool
	}{
		{"owner of property", owner, property, true, true},
		{"other owner", otherOwner, property, false, false},
		{"manager of property", manager, property, true, false},
		{"manager matching owner id", &common.Actor{ID: 1, Role: common.AdminRoleManager.ToString()}, property, false, false},
		{"manager of unmanaged property", &common.Actor{ID: 0, Role: common.AdminRoleManager.ToString()}, unmanaged, false, false},
		{"tenant", &common.Actor{ID: 1, Role: common.ActorRoleTenant}, property, false, false},
		{"missing property", owner, nil, false, false},
		{"missing actor", nil, property, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.property.AccessibleBy(tc.actor); got != tc.wantAccess {
				t.Fatalf("AccessibleBy() = %v, want %v", got, tc.wantAccess)
			}
			if got := tc.property.OwnedBy(tc.actor); got != tc.wantOwnedBy {
				t.Fatalf("OwnedBy() = %v, want %v", got, tc.wantOwnedBy)
			}
		})
	}
}
