package cpaas

import (
	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/config"
	"github.com/patrickishaf/property-mgt/internal/net"
)

type Client struct {
	cfg  *config.Config
	exec *common.CronExecutor
	http *net.HTTPClient
}

func NewClient(cfg *config.Config, exec *common.CronExecutor) *Client {
	headers := make(map[string]string)
	headers["Authorization"] = cfg.Cpaas.APIKey
	headers["Accept"] = "application/json"
	headers["Content-Type"] = "application/json"

	return &Client{
		cfg:  cfg,
		exec: exec,
		http: net.NewHTTPClient(cfg.Cpaas.APIBaseURL, headers),
	}
}

func (this *Client) sendEmail(emailType EmailType, recipient EmailRecipient, payload map[string]string) error {
	data := map[string]any{
		"template_key": this.templateForEmailType(emailType),
		"merge_info":   payload,
		"from": map[string]string{
			"address": this.cfg.Cpaas.SenderAddress,
			"name":    this.cfg.App.AppName,
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

func (this *Client) sendSMS() {}

func (this *Client) templateForEmailType(emailType EmailType) string {
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
