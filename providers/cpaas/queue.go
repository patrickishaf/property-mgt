package cpaas

import (
	"encoding/json"
	"time"

	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/config"
	"github.com/patrickishaf/property-mgt/internal/db/model"
	"github.com/patrickishaf/property-mgt/internal/db/repository"
	"gorm.io/gorm"
)

// QueueInviteEmail creates an invite token for the actor and queues the invite email
// inside the given transaction. It returns the set-password link so callers can log
// it in development.
func QueueInviteEmail(cfg *config.AppConfig, jobRepo *repository.JobRepository, trx *gorm.DB, actor common.Actor, fullName string, emailType EmailType, subject string) (string, error) {
	inviteToken, err := common.GenerateInviteJWT(*cfg, actor)
	if err != nil {
		return "", err
	}
	link := cfg.WebAppURL + "/set-password?token=" + inviteToken
	payload, err := json.Marshal(EmailJob{
		Payload:   map[string]string{"firstName": actor.FirstName, "link": link},
		Recipient: EmailRecipient{Address: actor.Email, Name: fullName},
		Subject:   subject,
		Type:      emailType,
	})
	if err != nil {
		return "", err
	}
	now := time.Now()
	_, err = jobRepo.SaveDefaultJobTRX(&model.Job{
		DueDate:         &now,
		LastScheduledAt: now,
		Payload:         string(payload),
		Status:          common.ProcessingStatusPending.String(),
		Type:            model.JobTypeEmail.String(),
	}, trx)
	if err != nil {
		return "", err
	}
	return link, nil
}
