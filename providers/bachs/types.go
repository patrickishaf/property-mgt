package bachs

import "errors"

var ErrNotImplemented = errors.New("bachs: not implemented")

type CheckoutStatus string

const (
	OpenCheckoutStatus      CheckoutStatus = "open"
	CompletedCheckoutStatus CheckoutStatus = "completed"
	ExpiredCheckoutStatus   CheckoutStatus = "expired"
	CancelledCheckoutStatus CheckoutStatus = "cancelled"
)

const (
	PriceTypeFixed  = "fixed"
	PriceTypeCustom = "custom"
	PriceTypeFree   = "free"
)

type Pricing struct {
	PriceType string `json:"price_type"`
	Amount    string `json:"amount,omitempty"`
	Currency  string `json:"currency"`
}

// Customer identifies the payer, either by an existing Bachs customer ID or by
// contact details.
type Customer struct {
	CustomerID  string `json:"customer_id,omitempty"`
	Email       string `json:"email,omitempty"`
	Name        string `json:"name,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

// CreateCheckoutSessionBody is the request body for POST /v1/checkout-sessions.
type CreateCheckoutSessionBody struct {
	Pricing           Pricing           `json:"pricing"`
	CancelURL         string            `json:"cancel_url"`
	SuccessURL        string            `json:"success_url"`
	Customer          Customer          `json:"customer"`
	Metadata          map[string]string `json:"metadata"`
	Reference         string            `json:"reference"`
	SavePaymentMethod bool              `json:"save_payment_method"`
}

// CheckoutRes is a checkout session returned by Bachs.
type CheckoutRes struct {
	CheckoutID        string         `json:"checkout_id"`
	CheckoutURL       string         `json:"checkout_url"`
	Status            CheckoutStatus `json:"status"`
	ExpiresAt         string         `json:"expires_at"`
	CreatedAt         string         `json:"created_at"`
	Reference         string         `json:"reference"`
	PlatformFee       *string        `json:"platform_fee"`
	DestinationAmount string         `json:"destination_amount"`
}

const (
	EventCheckoutCompleted = "checkout.completed"
	EventCheckoutExpired   = "checkout.expired"
)

// WebhookEvent is a webhook delivery. Only the fields the app uses are mapped.
type WebhookEvent struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	CreatedAt string          `json:"created_at"`
	Data      WebhookCheckout `json:"data"`
}

type WebhookCheckout struct {
	CheckoutID    string            `json:"checkout_id"`
	Status        CheckoutStatus    `json:"status"`
	PaymentStatus *string           `json:"payment_status"`
	Amount        string            `json:"amount"`
	Currency      string            `json:"currency"`
	Reference     string            `json:"reference"`
	Metadata      map[string]string `json:"metadata"`
	CompletedAt   *string           `json:"completed_at"`
	ExpiresAt     string            `json:"expires_at"`
}
