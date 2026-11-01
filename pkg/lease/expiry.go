package lease

import (
	"errors"
	"time"

	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/internal/db/model"
	"gorm.io/gorm"
)

// expiryCutoff is the start of the given day in UTC. Leases are stored with
// date-only end dates, so a lease ending on day D stays active through D and is
// ended once the cutoff passes it.
func expiryCutoff(now time.Time) time.Time {
	y, m, d := now.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// RegisterExpiryJob schedules a daily job that ends active leases whose end date
// has passed.
func (this *service) RegisterExpiryJob(scheduler *common.Scheduler) {
	_, err := scheduler.Cron().AddFunc("@daily", func() {
		this.expireLeases(time.Now())
	})
	if err != nil {
		this.logger.Error("failed to schedule lease expiry job", "err", err.Error())
	}
}

// expireLeases ends every active lease that ended before today and returns how
// many were ended. Each lease is ended in its own transaction; failures are
// logged and skipped.
func (this *service) expireLeases(now time.Time) int {
	logger := this.logger.With("caller", "lease.service.expireLeases")
	ids, err := this.leaseRepo.FindActiveEndingBefore(expiryCutoff(now))
	if err != nil {
		logger.Error("failed to find expired leases", "err", err.Error())
		return 0
	}
	ended := 0
	for _, id := range ids {
		err := this.db.Transaction(func(trx *gorm.DB) error {
			lease, err := this.leaseRepo.FindOneByIdForUpdateTRX(id, trx)
			if err != nil {
				return err
			}
			if lease == nil {
				return nil
			}
			return this.endLeaseTRX(lease, model.LeaseStatusEnded, now, trx)
		})
		switch {
		case err == nil:
			ended++
		case errors.Is(err, errNotActive):
			// Ended by someone else since the lookup.
		default:
			logger.Error("failed to end expired lease", "lease_id", id, "err", err.Error())
		}
	}
	if ended > 0 {
		logger.Info("ended expired leases", "count", ended)
	}
	return ended
}

// runExpiry runs the expiry job immediately; development only.
func (this *service) runExpiry() common.ServiceResponse {
	return common.ServiceResponse{StatusCode: 200, Data: map[string]int{"ended": this.expireLeases(time.Now())}}
}
