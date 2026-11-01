package payment

import (
	"errors"
	"net/http"
	"time"

	"github.com/patrickishaf/property-mgt/common"
	"github.com/patrickishaf/property-mgt/internal/db/model"
	"github.com/patrickishaf/property-mgt/providers/bachs"
)

// Webhook outcomes. Anything a retry can't change is acknowledged with 200 so
// Bachs stops redelivering it; only failures on our side (500) invite a retry.
const (
	webhookApplied   = "payment applied"
	webhookExpired   = "checkout marked expired"
	webhookDuplicate = "already processed"
	webhookIgnored   = "ignored"
)

func webhookAck(outcome string) common.ServiceResponse {
	return common.ServiceResponse{StatusCode: 200, Data: map[string]string{"outcome": outcome}}
}

// verifyBachsWebhook reports whether a delivery provably came from Bachs: its
// signature header must sign the raw body with BACHS_WH_SECRET.
func (this *service) verifyBachsWebhook(header http.Header, body []byte) bool {
	return this.bachsClient.VerifyWebhook(header, body)
}

// handleBachsEvent applies a verified Bachs webhook to the checkout payment it
// refers to, found by the checkout's reference.
func (this *service) handleBachsEvent(event bachs.WebhookEvent) common.ServiceResponse {
	logger := this.logger.With("caller", "payment.service.handleBachsEvent", "event_id", event.ID,
		"event_type", event.Type, "reference", event.Data.Reference)
	if kind := event.Data.Metadata["kind"]; kind != "" && kind != metadataKind {
		return webhookAck(webhookIgnored)
	}
	if event.Type != bachs.EventCheckoutCompleted && event.Type != bachs.EventCheckoutExpired {
		logger.Info("ignoring unhandled bachs event")
		return webhookAck(webhookIgnored)
	}

	payment, err := this.paymentRepo.FindOneByReference(event.Data.Reference)
	if err != nil {
		logger.Error("failed to find payment for webhook", "err", err.Error())
		return serverError()
	}
	if payment == nil || payment.Source != model.PaymentSourceCheckout {
		logger.Warn("bachs webhook for an unknown checkout")
		return webhookAck(webhookIgnored)
	}
	if id := event.Data.Metadata["payment_id"]; id != "" && id != formatID(payment.ID) {
		logger.Error("bachs webhook metadata names a different payment", "metadata_payment_id", id, "payment_id", payment.ID)
		return webhookAck(webhookIgnored)
	}

	if event.Type == bachs.EventCheckoutExpired {
		if payment.Status != model.PaymentStatusUnpaid {
			return webhookAck(webhookDuplicate)
		}
		if err := this.paymentRepo.UpdateTRX(payment.ID, map[string]any{"status": model.PaymentStatusExpired}, this.db); err != nil {
			logger.Error("failed to mark checkout expired", "err", err.Error())
			return serverError()
		}
		return webhookAck(webhookExpired)
	}

	if payment.Status == model.PaymentStatusPaid || payment.Status == model.PaymentStatusPartiallyPaid {
		return webhookAck(webhookDuplicate)
	}
	paid := event.Data.PaymentStatus != nil && *event.Data.PaymentStatus == "paid"
	if !paid || event.Data.Status != bachs.CompletedCheckoutStatus {
		logger.Warn("completed checkout is not paid", "status", event.Data.Status)
		return webhookAck(webhookIgnored)
	}
	// The provider must have charged exactly what the checkout asked for.
	if event.Data.Amount != formatAmount(payment.Amount) || event.Data.Currency != payment.Currency {
		logger.Error("bachs charged a different amount than the checkout; not applied",
			"charged", event.Data.Amount+" "+event.Data.Currency,
			"expected", formatAmount(payment.Amount)+" "+payment.Currency)
		return webhookAck(webhookIgnored)
	}

	// Use when the provider completed it, so a late or retried delivery isn't
	// mistaken for a payment after the checkout expired.
	paidAt := time.Now()
	if event.Data.CompletedAt != nil {
		if at, err := time.Parse(time.RFC3339Nano, *event.Data.CompletedAt); err == nil {
			paidAt = at
		}
	}
	_, err = this.settleCheckout(payment.Reference, paidAt)
	switch {
	case err == nil:
		return webhookAck(webhookApplied)
	case errors.Is(err, errCheckoutNotUnpaid):
		return webhookAck(webhookDuplicate)
	case errors.Is(err, errCheckoutExpired), errors.Is(err, errAmountExceedsOutstanding):
		// Money was taken but can't be applied as-is (e.g. the balance was
		// paid another way meanwhile); someone has to sort it out by hand.
		logger.Error("paid checkout could not be applied; needs manual attention", "err", err.Error())
		return webhookAck(webhookIgnored)
	default:
		logger.Error("failed to apply paid checkout", "err", err.Error())
		return serverError()
	}
}
