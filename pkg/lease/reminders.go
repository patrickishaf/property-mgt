package lease

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/internal/db/model"
	"github.com/patrickishaf/property-mgt/pkg/messaging"
)

// reminderMessenger sends a message on one channel; the messaging service
// implements it.
type reminderMessenger interface {
	Send(from messaging.Sender, channel string, templateKey string, subject string, body string, recipients []messaging.Recipient) (uint, error)
}

const (
	reminderSenderName = "Aventra reminders"
	// Overdue rent and expired leases are re-sent weekly; an expired lease
	// gets at most four texts, so a tenant who moved out isn't texted forever.
	reminderRepeatEvery = 7 * 24 * time.Hour
	maxExpiredReminders = 4
	// Expired leases are followed up for this long after their end date.
	expiredReminderWindow = 4 * 7 * 24 * time.Hour
)

// RegisterReminderJob schedules the daily tenant reminders for 9am Lagos time,
// so texts don't arrive at night.
func (this *service) RegisterReminderJob(scheduler *common.Scheduler, messenger reminderMessenger) {
	this.messenger = messenger
	_, err := scheduler.Cron().AddFunc("CRON_TZ=Africa/Lagos 0 9 * * *", func() {
		this.sendReminders(time.Now())
	})
	if err != nil {
		this.logger.Error("failed to schedule lease reminder job", "err", err.Error())
	}
}

// expiringMilestone is the reminder due for a lease ending in daysLeft days:
// the nearest of 30, 14 and 7 days out that has been reached.
func expiringMilestone(daysLeft int) (string, bool) {
	switch {
	case daysLeft < 0 || daysLeft > 30:
		return "", false
	case daysLeft <= 7:
		return model.LeaseReminderExpiring7, true
	case daysLeft <= 14:
		return model.LeaseReminderExpiring14, true
	default:
		return model.LeaseReminderExpiring30, true
	}
}

// dueAgain reports whether a repeating reminder last sent at last is due now.
func dueAgain(last *model.LeaseReminder, now time.Time) bool {
	return last == nil || now.Sub(last.SentAt) >= reminderRepeatEvery
}

// formatNaira renders an amount as ₦1,050,000 (or ₦1,050.50 with kobo).
func formatNaira(amount float64) string {
	amount = math.Round(amount*100) / 100
	whole := int64(amount)
	digits := strconv.FormatInt(whole, 10)
	var b strings.Builder
	for i, d := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(d)
	}
	if kobo := int64(math.Round((amount - float64(whole)) * 100)); kobo > 0 {
		return fmt.Sprintf("₦%s.%02d", b.String(), kobo)
	}
	return "₦" + b.String()
}

type reminderCounts struct {
	Expiring int `json:"expiring"`
	Overdue  int `json:"overdue"`
	Expired  int `json:"expired"`
	Skipped  int `json:"skipped"`
}

// reminderTarget is a lease being reminded about, with who to send it to and
// on whose behalf.
type reminderTarget struct {
	lease     model.Lease
	recipient messaging.Recipient
	ownerID   uint
}

// sendReminders sends every reminder due today and returns how many went out:
//   - leases ending in 30, 14 or 7 days: one in-app message per milestone;
//   - leases with overdue rent: an SMS when it becomes overdue, then weekly;
//   - leases that expired (ended on their end date) in the last four weeks with
//     no newer lease for the tenant: an SMS on expiry, then weekly, at most four.
//
// SMS are queued for the SMS worker. Failures are logged and skipped, and a
// reminder is only recorded once it has been sent.
func (this *service) sendReminders(now time.Time) reminderCounts {
	var counts reminderCounts
	if this.messenger == nil {
		this.logger.Error("lease reminders have no messenger configured")
		return counts
	}
	today := expiryCutoff(now)
	this.sendExpiringReminders(now, today, &counts)
	this.sendOverdueReminders(now, &counts)
	this.sendExpiredReminders(now, today, &counts)
	return counts
}

func (this *service) sendExpiringReminders(now time.Time, today time.Time, counts *reminderCounts) {
	logger := this.logger.With("caller", "lease.service.sendExpiringReminders")
	leases, err := this.leaseRepo.FindByStatusEndingBetween(model.LeaseStatusActive, today, today.AddDate(0, 0, 31))
	if err != nil {
		logger.Error("failed to find expiring leases", "err", err.Error())
		return
	}
	for _, target := range this.reminderTargets(leases, logger) {
		daysLeft := int(math.Round(target.lease.EndDate.Sub(today).Hours() / 24))
		kind, ok := expiringMilestone(daysLeft)
		if !ok {
			continue
		}
		last, err := this.reminderRepo.FindLatest(target.lease.ID, kind)
		if err != nil {
			logger.Error("failed to check reminders", "lease_id", target.lease.ID, "err", err.Error())
			continue
		}
		if last != nil {
			continue
		}
		ends := target.lease.EndDate.Format("2 Jan 2006")
		in := fmt.Sprintf("in %d days", daysLeft)
		if daysLeft == 1 {
			in = "tomorrow"
		} else if daysLeft == 0 {
			in = "today"
		}
		this.sendReminder(target, kind, model.MessageChannelInApp, "", "Your lease ends on "+ends,
			fmt.Sprintf("Dear {name}, your lease for {unit} at {property} ends on %s (%s). Please contact management if you'd like to renew.", ends, in),
			now, &counts.Expiring, counts)
	}
}

func (this *service) sendOverdueReminders(now time.Time, counts *reminderCounts) {
	logger := this.logger.With("caller", "lease.service.sendOverdueReminders")
	installments, err := this.installmentRepo.FindOpen(0, 0, 0, 0)
	if err != nil {
		logger.Error("failed to find open installments", "err", err.Error())
		return
	}
	unscheduled, err := this.leaseRepo.FindOpenUnscheduled(0, 0, 0, 0)
	if err != nil {
		logger.Error("failed to find open leases", "err", err.Error())
		return
	}
	// The same rule as the due-dates list: past its due date with money owed.
	overdue := map[uint]float64{}
	for _, item := range buildDueItems(installments, unscheduled, now, "overdue") {
		overdue[item.LeaseID] += item.Amount
	}
	ids := make([]uint, 0, len(overdue))
	for id := range overdue {
		ids = append(ids, id)
	}
	leases, err := this.leaseRepo.FindByIds(ids)
	if err != nil {
		logger.Error("failed to load overdue leases", "err", err.Error())
		return
	}
	for _, target := range this.reminderTargets(leases, logger) {
		last, err := this.reminderRepo.FindLatest(target.lease.ID, model.LeaseReminderOverdueRent)
		if err != nil {
			logger.Error("failed to check reminders", "lease_id", target.lease.ID, "err", err.Error())
			continue
		}
		if !dueAgain(last, now) {
			continue
		}
		this.sendReminder(target, model.LeaseReminderOverdueRent, model.MessageChannelSMS, "payment_reminder", "Rent overdue",
			fmt.Sprintf("Dear {name}, rent of %s for {unit} at {property} is overdue. Please pay as soon as you can.", formatNaira(overdue[target.lease.ID])),
			now, &counts.Overdue, counts)
	}
}

func (this *service) sendExpiredReminders(now time.Time, today time.Time, counts *reminderCounts) {
	logger := this.logger.With("caller", "lease.service.sendExpiredReminders")
	// Ended (not terminated) leases whose end date has passed: expired, either
	// by the expiry job or ended by management at the end of the term.
	leases, err := this.leaseRepo.FindByStatusEndingBetween(model.LeaseStatusEnded, today.Add(-expiredReminderWindow), today)
	if err != nil {
		logger.Error("failed to find expired leases", "err", err.Error())
		return
	}
	for _, target := range this.reminderTargets(leases, logger) {
		renewed, err := this.leaseRepo.FindActiveByTenant(uint(target.lease.TenantID))
		if err != nil {
			logger.Error("failed to check for a newer lease", "lease_id", target.lease.ID, "err", err.Error())
			continue
		}
		if renewed != nil {
			continue
		}
		sent, err := this.reminderRepo.Count(target.lease.ID, model.LeaseReminderExpired)
		if err != nil {
			logger.Error("failed to count reminders", "lease_id", target.lease.ID, "err", err.Error())
			continue
		}
		last, err := this.reminderRepo.FindLatest(target.lease.ID, model.LeaseReminderExpired)
		if err != nil {
			logger.Error("failed to check reminders", "lease_id", target.lease.ID, "err", err.Error())
			continue
		}
		if sent >= maxExpiredReminders || !dueAgain(last, now) {
			continue
		}
		this.sendReminder(target, model.LeaseReminderExpired, model.MessageChannelSMS, "", "Lease expired",
			fmt.Sprintf("Dear {name}, your lease for {unit} at {property} ended on %s. Please contact management about renewing.", target.lease.EndDate.Format("2 Jan 2006")),
			now, &counts.Expired, counts)
	}
}

// reminderTargets pairs leases with their tenant's contact details and their
// property's owner, skipping any whose tenant or property is gone.
func (this *service) reminderTargets(leases []model.Lease, logger interface {
	Error(msg string, args ...any)
}) []reminderTarget {
	if len(leases) == 0 {
		return nil
	}
	tenantIDs := make([]uint, 0, len(leases))
	propertyIDs := make([]uint, 0, len(leases))
	for _, l := range leases {
		tenantIDs = append(tenantIDs, uint(l.TenantID))
		propertyIDs = append(propertyIDs, uint(l.PropertyID))
	}
	tenants, err := this.tenantRepo.FindByIds(tenantIDs)
	if err != nil {
		logger.Error("failed to load tenants", "err", err.Error())
		return nil
	}
	properties, err := this.propertyRepo.FindByIds(propertyIDs)
	if err != nil {
		logger.Error("failed to load properties", "err", err.Error())
		return nil
	}
	tenantByID := map[uint]model.Tenant{}
	for _, t := range tenants {
		tenantByID[t.ID] = t
	}
	propertyByID := map[uint]model.Property{}
	for _, p := range properties {
		propertyByID[p.ID] = p
	}
	targets := make([]reminderTarget, 0, len(leases))
	for _, l := range leases {
		t, okT := tenantByID[uint(l.TenantID)]
		p, okP := propertyByID[uint(l.PropertyID)]
		if !okT || !okP {
			continue
		}
		targets = append(targets, reminderTarget{
			lease:   l,
			ownerID: uint(p.OwnerID),
			recipient: messaging.Recipient{
				Email:        t.Email,
				FullName:     strings.TrimSpace(t.FirstName + " " + t.LastName),
				PhoneNumber:  t.PhoneNumber,
				PropertyID:   p.ID,
				PropertyName: p.Name,
				TenantID:     t.ID,
				UnitNumber:   l.UnitNumber,
			},
		})
	}
	return targets
}

// sendReminder sends one reminder on the owner's behalf (so it shows in their
// sent messages) and records it. SMS reminders need a phone number.
func (this *service) sendReminder(target reminderTarget, kind string, channel string, templateKey string, subject string, body string,
	now time.Time, sent *int, counts *reminderCounts,
) {
	logger := this.logger.With("caller", "lease.service.sendReminder", "lease_id", target.lease.ID, "kind", kind)
	if channel == model.MessageChannelSMS && strings.TrimSpace(target.recipient.PhoneNumber) == "" {
		logger.Info("tenant has no phone number; reminder skipped")
		counts.Skipped++
		return
	}
	ownerID := target.ownerID
	messageID, err := this.messenger.Send(messaging.Sender{ID: &ownerID, Name: reminderSenderName}, channel, templateKey,
		subject, body, []messaging.Recipient{target.recipient})
	if err != nil {
		logger.Error("failed to send reminder", "err", err.Error())
		counts.Skipped++
		return
	}
	if err := this.reminderRepo.SaveOne(&model.LeaseReminder{LeaseID: target.lease.ID, Kind: kind, MessageID: messageID, SentAt: now}); err != nil {
		logger.Error("failed to record reminder", "err", err.Error())
	}
	*sent++
}

// runReminders sends today's reminders now; development only.
func (this *service) runReminders() common.ServiceResponse {
	return common.ServiceResponse{StatusCode: 200, Data: this.sendReminders(time.Now())}
}
