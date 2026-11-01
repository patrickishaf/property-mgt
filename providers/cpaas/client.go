package cpaas

import (
	"log/slog"

	"github.com/patrickishaf/property-mgt/config"
	"github.com/patrickishaf/property-mgt/internal/net"
)

type client struct {
	cfg  *config.CpaasConfig
	http *net.HTTPClient
}

func NewClient(cfg *config.CpaasConfig) *client {
	headers := make(map[string]string)
	headers["Authorization"] = cfg.APIKey
	headers["Accept"] = "application/json"
	headers["Content-Type"] = "application/json"

	return &client{
		cfg:  cfg,
		http: net.NewHTTPClient(cfg.APIBaseURL, headers),
	}
}

// sendEmail sends a queued email using the cpaas template for its type, with the
// job's subject and merge fields.
func (this *client) sendEmail(email EmailJob) error {
	data := map[string]any{
		"template_key": this.templateForEmailType(email.Type),
		"merge_info":   email.Payload,
		"from": map[string]string{
			"address": this.cfg.SenderAddress,
			"name":    this.cfg.SenderName,
		},
		"to":      this.toEmailReceivers(email.Recipients),
		"subject": email.Subject,
	}
	_, err := this.http.Post("/v1.1/email/template", data)
	if err != nil {
		return err
	}
	return nil
}

func (this *client) toEmailReceivers(recipients []EmailRecipient) []emailReceiver {
	receivers := make([]emailReceiver, 0, len(recipients))
	for _, r := range recipients {
		receivers = append(receivers, emailReceiver{
			"email_address": map[string]string{
				"address": r.Address,
				"name":    r.Name,
			},
		})
	}
	return receivers
}

// sendSMS sends a text message.
//
// TODO: call the SMS provider once its endpoint and payload format are known.
// Until then this only logs the message.
func (this *client) sendSMS(sms SMSJob) error {
	slog.Info("cpaas sms (stub)", "to", sms.PhoneNumber, "message_recipient_id", sms.MessageRecipientID)
	return nil
}

func (this *client) templateForEmailType(emailType EmailType) string {
	switch emailType {
	case EmailTypeGeneralNotice:
		return "general_notice"
	case EmailTypePaymentReminder:
		return "payment_reminder"
	case EmailTypeRentIncrease:
		return "rent_increase"
	case EmailTypeScheduledMaintenance:
		return "scheduled_maintenance"
	default:
		return string(emailType)
	}
}
