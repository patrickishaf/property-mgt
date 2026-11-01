package payment

import (
	"reflect"
	"testing"
	"time"

	"github.com/patrickishaf/property-mgt/internal/db/repository"
)

func TestYearBounds(t *testing.T) {
	now := time.Date(2026, 12, 31, 23, 30, 0, 0, time.FixedZone("WAT", 3600)) // 22:30 UTC
	start, end := yearBounds(0, now)
	if start != time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) || end != time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC) {
		t.Fatalf("default year bounds = %v..%v", start, end)
	}
	start, end = yearBounds(2025, now)
	if start.Year() != 2025 || end.Year() != 2026 {
		t.Fatalf("2025 bounds = %v..%v", start, end)
	}
}

func TestPortfolioFromTotals(t *testing.T) {
	res := portfolioFromTotals(
		repository.DueTotals{Expected: 800, OpenCount: 2, OverdueCount: 1, OverdueAmount: 400},
		repository.DueTotals{Expected: 1200, OpenCount: 1, OverdueCount: 1, OverdueAmount: 1200},
		500, 2,
	)
	want := portfolioRes{
		ActiveTenantCount:       2,
		OpenInstallmentCount:    3,
		OutstandingRentBalance:  1500,
		OverdueInstallmentCount: 2,
		PercentageRentCollected: 25,
		TotalOverdueAmount:      1600,
		TotalRentCollected:      500,
		TotalRentExpected:       2000,
	}
	if res != want {
		t.Fatalf("portfolioFromTotals() = %+v, want %+v", res, want)
	}

	over := portfolioFromTotals(repository.DueTotals{Expected: 100}, repository.DueTotals{}, 300, 0)
	if over.OutstandingRentBalance != 0 || over.PercentageRentCollected != 300 {
		t.Fatalf("collected above expected should clamp outstanding to 0: %+v", over)
	}
	if empty := portfolioFromTotals(repository.DueTotals{}, repository.DueTotals{}, 0, 0); empty.PercentageRentCollected != 0 {
		t.Fatalf("no expected rent should give 0%%, got %v", empty.PercentageRentCollected)
	}
}

func TestParseStatuses(t *testing.T) {
	cases := []struct {
		in     string
		want   []string
		wantOK bool
	}{
		{"", nil, true},
		{"paid", []string{"paid"}, true},
		{"paid, partially_paid,paid", []string{"paid", "partially_paid"}, true},
		{"paid,bogus", nil, false},
	}
	for _, tc := range cases {
		got, ok := parseStatuses(tc.in)
		if ok != tc.wantOK || !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("parseStatuses(%q) = %v, %v; want %v, %v", tc.in, got, ok, tc.want, tc.wantOK)
		}
	}
}
