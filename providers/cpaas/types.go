package cpaas

type emailReceiver map[string]map[string]string

type EmailJob struct {
	// MessageRecipientID links the job to a messaging recipient so the worker
	// can record delivery; zero for emails that aren't messages (e.g. invites).
	MessageRecipientID uint `json:",omitempty"`
	Payload            map[string]string
	// Recipients are all sent in the request's "to" list.
	Recipients []EmailRecipient
	Subject    string
	Type       EmailType
}

type SMSJob struct {
	Body               string
	MessageRecipientID uint `json:",omitempty"`
	PhoneNumber        string
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
	EmailTypeManagerInvite        EmailType = "manager_invite"
)
