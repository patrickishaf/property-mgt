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
	_, err = QueueEmail(jobRepo, trx, EmailJob{
		Payload:    map[string]string{"firstName": actor.FirstName, "link": link},
		Recipients: []EmailRecipient{{Address: actor.Email, Name: fullName}},
		Subject:    subject,
		Type:       emailType,
	})
	if err != nil {
		return "", err
	}
	return link, nil
}

// QueueEmail queues an email job inside the given transaction for the cpaas
// worker to send.
func QueueEmail(jobRepo *repository.JobRepository, trx *gorm.DB, email EmailJob) (*model.Job, error) {
	return queueJob(jobRepo, trx, model.JobTypeEmail, email)
}

// QueueSMS queues an SMS job inside the given transaction for the cpaas worker
// to send.
func QueueSMS(jobRepo *repository.JobRepository, trx *gorm.DB, sms SMSJob) (*model.Job, error) {
	return queueJob(jobRepo, trx, model.JobTypeSMS, sms)
}

func queueJob(jobRepo *repository.JobRepository, trx *gorm.DB, jobType model.JobType, payload any) (*model.Job, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	return jobRepo.SaveDefaultJobTRX(&model.Job{
		DueDate:         &now,
		LastScheduledAt: now,
		Payload:         string(data),
		Status:          common.ProcessingStatusPending.String(),
		Type:            jobType.String(),
	}, trx)
}
