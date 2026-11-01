package lease

import (
	"testing"
	"time"

	"github.com/patrickishaf/property-mgt/internal/db/model"
)

func ptr[T any](v T) *T {
	return &v
}

func TestCheckLeaseUpdate(t *testing.T) {
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	active := model.Lease{
		AmountPaid:           400,
		ExpectedRent:         1200,
		MinimumPartialAmount: 100,
		StartDate:            start,
		Status:               model.LeaseStatusActive,
	}
	ended := active
	ended.Status = model.LeaseStatusEnded

	cases := []struct {
		name     string
		lease    model.Lease
		dto      updateLeaseReq
		open     int64
		wantCode int
	}{
		{"partial settings on active lease", active, updateLeaseReq{AllowPartialPayment: ptr(true), MinimumPartialAmount: ptr(200.0)}, 0, 0},
		{"partial settings on ended lease", ended, updateLeaseReq{AllowPartialPayment: ptr(false)}, 0, 0},
		{"end date on ended lease", ended, updateLeaseReq{EndDate: ptr("2027-12-31")}, 0, 409},
		{"terminate ended lease", ended, updateLeaseReq{Status: ptr(model.LeaseStatusTerminated)}, 0, 409},
		{"end date before start", active, updateLeaseReq{EndDate: ptr("2026-09-30")}, 0, 400},
		{"end date in the past but after start", active, updateLeaseReq{EndDate: ptr("2026-10-15")}, 0, 0},
		{"rent change with unpaid installments", active, updateLeaseReq{ExpectedRent: ptr(1500.0)}, 2, 409},
		{"rent below amount paid", active, updateLeaseReq{ExpectedRent: ptr(300.0)}, 0, 400},
		{"rent equal to amount paid", active, updateLeaseReq{ExpectedRent: ptr(400.0)}, 0, 0},
		{"rent change below current minimum", active, updateLeaseReq{ExpectedRent: ptr(450.0), MinimumPartialAmount: ptr(500.0)}, 0, 400},
		{"minimum above rent", active, updateLeaseReq{MinimumPartialAmount: ptr(1300.0)}, 0, 400},
		{"terminate active lease", active, updateLeaseReq{Status: ptr(model.LeaseStatusTerminated)}, 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := checkLeaseUpdate(tc.lease, tc.dto, tc.open)
			got := 0
			if res != nil {
				got = res.StatusCode
			}
			if got != tc.wantCode {
				t.Fatalf("checkLeaseUpdate() status = %d, want %d (%+v)", got, tc.wantCode, res)
			}
		})
	}
}

func TestExpiryCutoff(t *testing.T) {
	lagos := time.FixedZone("WAT", 3600)
	now := time.Date(2026, 11, 1, 0, 30, 0, 0, lagos) // 2026-10-31 23:30 UTC
	want := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	if got := expiryCutoff(now); !got.Equal(want) {
		t.Fatalf("expiryCutoff() = %v, want %v", got, want)
	}

	endsToday := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	if endsToday.Before(expiryCutoff(now)) {
		t.Fatal("a lease ending today must not be expired yet")
	}
	endedYesterday := time.Date(2026, 10, 30, 0, 0, 0, 0, time.UTC)
	if !endedYesterday.Before(expiryCutoff(now)) {
		t.Fatal("a lease that ended yesterday must be expired")
	}
}
