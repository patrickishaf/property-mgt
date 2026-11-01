package auth

import (
	"reflect"
	"testing"
)

func TestProfileFields(t *testing.T) {
	str := func(s string) *string { return &s }
	cases := []struct {
		name string
		dto  updateUserReq
		want map[string]any
		ok   bool
	}{
		{"nothing sent", updateUserReq{}, nil, false},
		{"name only, trimmed", updateUserReq{FirstName: str("  Ada "), LastName: str("Obi")},
			map[string]any{"first_name": "Ada", "last_name": "Obi"}, true},
		{"phone only", updateUserReq{PhoneNumber: str("0803 555 0100")},
			map[string]any{"phone_number": "0803 555 0100"}, true},
		{"blank value rejected", updateUserReq{FirstName: str("Ada"), LastName: str("   ")}, nil, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := profileFields(c.dto)
			if ok != c.ok || (c.ok && !reflect.DeepEqual(got, c.want)) {
				t.Fatalf("profileFields = %v, %v; want %v, %v", got, ok, c.want, c.ok)
			}
		})
	}
}
