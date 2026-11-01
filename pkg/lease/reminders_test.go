package lease

import (
	"testing"
	"time"

	"github.com/patrickishaf/property-mgt/internal/db/model"
)

func TestExpiringMilestone(t *testing.T) {
	cases := map[int]string{
		31: "", -1: "",
		30: model.LeaseReminderExpiring30, 15: model.LeaseReminderExpiring30,
		14: model.LeaseReminderExpiring14, 8: model.LeaseReminderExpiring14,
		7: model.LeaseReminderExpiring7, 0: model.LeaseReminderExpiring7,
	}
	for days, want := range cases {
		got, ok := expiringMilestone(days)
		if ok != (want != "") || got != want {
			t.Errorf("expiringMilestone(%d) = %q, %v; want %q", days, got, ok, want)
		}
	}
}

func TestDueAgain(t *testing.T) {
	now := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
	if !dueAgain(nil, now) {
		t.Error("never sent should be due")
	}
	if dueAgain(&model.LeaseReminder{SentAt: now.AddDate(0, 0, -6)}, now) {
		t.Error("sent 6 days ago should not be due")
	}
	if !dueAgain(&model.LeaseReminder{SentAt: now.AddDate(0, 0, -7)}, now) {
		t.Error("sent 7 days ago should be due")
	}
}

func TestFormatNaira(t *testing.T) {
	cases := map[float64]string{0: "₦0", 950: "₦950", 1050000: "₦1,050,000", 1234.5: "₦1,234.50", 100000.004: "₦100,000"}
	for amount, want := range cases {
		if got := formatNaira(amount); got != want {
			t.Errorf("formatNaira(%v) = %q, want %q", amount, got, want)
		}
	}
}
