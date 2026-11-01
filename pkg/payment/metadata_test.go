package payment

import (
	"reflect"
	"testing"

	"github.com/patrickishaf/property-mgt/internal/db/model"
)

func TestCheckoutMetadata(t *testing.T) {
	base := model.Payment{LeaseID: 12, TenantID: 7, PropertyID: 3}
	base.ID = 40
	common := map[string]string{
		"kind": "rent_payment", "schema_version": "1", "payment_id": "40",
		"lease_id": "12", "tenant_id": "7", "property_id": "3",
	}
	with := func(extra map[string]string) map[string]string {
		out := map[string]string{}
		for k, v := range common {
			out[k] = v
		}
		for k, v := range extra {
			out[k] = v
		}
		return out
	}
	cases := []struct {
		name         string
		purpose      string
		installments []uint
		want         map[string]string
	}{
		{"whole lease", model.PaymentPurposeLease, nil,
			with(map[string]string{"purpose": "lease", "target_type": "lease", "target_ids": "12"})},
		{"part of the lease", model.PaymentPurposePartial, nil,
			with(map[string]string{"purpose": "partial", "target_type": "lease", "target_ids": "12"})},
		{"installments", model.PaymentPurposeInstallments, []uint{31, 32},
			with(map[string]string{"purpose": "installments", "target_type": "installments", "target_ids": "31,32"})},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := base
			p.Purpose = c.purpose
			got := checkoutMetadata(p, c.installments).toMap()
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("metadata = %v, want %v", got, c.want)
			}
		})
	}
}
