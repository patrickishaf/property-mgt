package maintenance

import (
	"reflect"
	"testing"

	"github.com/patrickishaf/property-mgt/internal/db/model"
	"gorm.io/gorm"
)

func TestNotifyChannels(t *testing.T) {
	cases := []struct {
		name      string
		requested []string
		want      []string
	}{
		{"missing defaults to in-app", nil, []string{model.MessageChannelInApp}},
		{"empty defaults to in-app", []string{}, []string{model.MessageChannelInApp}},
		{"keeps requested order", []string{"sms", "email"}, []string{"sms", "email"}},
		{"removes duplicates", []string{"email", "in_app", "email"}, []string{"email", "in_app"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := notifyChannels(tc.requested); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("notifyChannels(%v) = %v, want %v", tc.requested, got, tc.want)
			}
		})
	}
}

func TestCheckLinkableComplaints(t *testing.T) {
	linked := uint(9)
	found := []model.Complaint{
		{Model: gorm.Model{ID: 1}, PropertyID: 5},
		{Model: gorm.Model{ID: 2}, PropertyID: 5},
		{Model: gorm.Model{ID: 3}, PropertyID: 6},
		{Model: gorm.Model{ID: 4}, PropertyID: 5, IssueID: &linked},
	}
	cases := []struct {
		name string
		ids  []uint
		want []uint
	}{
		{"all linkable", []uint{1, 2}, nil},
		{"wrong property", []uint{1, 3}, []uint{3}},
		{"already linked", []uint{4, 2}, []uint{4}},
		{"unknown complaint", []uint{99, 1}, []uint{99}},
		{"repeated invalid id reported once, sorted", []uint{99, 3, 99}, []uint{3, 99}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := checkLinkableComplaints(tc.ids, found, 5); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("checkLinkableComplaints(%v) = %v, want %v", tc.ids, got, tc.want)
			}
		})
	}
}
