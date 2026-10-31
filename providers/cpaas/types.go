package cpaas

type emailReceiver map[string]map[string]string

type EmailJob struct {
	Payload   map[string]string
	Recipient EmailRecipient
	Subject   string
	Type      EmailType
}

type EmailRecipient struct {
	Address string
	Name    string
}

type EmailType string

const (
	EmailTypeRentIncrease         EmailType = "rent_increase"
	EmailTypeScheduledMaintenance EmailType = "scheduled_maintenance"
	EmailTypePaymentReminder      EmailType = "payment_reminder"
	EmailTypeGeneralNotice        EmailType = "general_notice"
	EmailTypeTenantInvite         EmailType = "tenant_invite"
)
