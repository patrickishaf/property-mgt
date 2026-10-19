package cpaas

import (
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

func (this *client) sendEmail(emailType EmailType, recipient EmailRecipient, payload map[string]string) error {
	data := map[string]any{
		"template_key": this.templateForEmailType(emailType),
		"merge_info":   payload,
		"from": map[string]string{
			"address": this.cfg.SenderAddress,
			"name":    this.cfg.SenderName,
		},
		"to": []emailReceiver{
			{
				"email_address": map[string]string{
					"address": recipient.Address,
					"name":    recipient.name,
				},
			},
		},
		"subject": "Billing attempt failed",
	}
	_, err := this.http.Post("/v1.1/email/template", data)
	if err != nil {
		return err
	}
	return nil
}

func (this *client) sendSMS() {}

func (this *client) templateForEmailType(emailType EmailType) string {
	switch emailType {
	case EmailTypeGeneralNotice:
		return "general_notice"
	case EmailTypePaymentReminder:
		return "payment_reminder"
	case EmailTypeRentIncrease:
		return "rent_increase"
	case EmailTypeScheduledMaintenance:
		return "rent_increase"
	default:
		return string(emailType)
	}
}
