package bachs

type PriceType string

type CheckoutStatus string

const (
	OpenCheckoutStatus      CheckoutStatus = "open"
	CompletedCheckoutStatus CheckoutStatus = "completed"
	ExpiredCheckoutStatus   CheckoutStatus = "expired"
	CancelledCheckoutStatus CheckoutStatus = "cancelled"
)

type createCheckoutSessionBody struct {
	Pricing struct {
		Amount          string            `json:"amount"`
		Currency        string            `json:"currency"`
		PresetAmount    string            `json:"preset_amount"`
		PriceType       string            `json:"price_type"`
		CurrencyOptions map[string]string `json:"currency_options"`
	} `json:"pricing"`
	BillingCurrency    string   `json:"billing_currency"`
	PaymentMethodTypes []string `json:"payment_method_types"`
	CancelURL          string   `json:"cancel_url"`
	SuccessURL         string   `json:"success_url"`
	Customer           struct {
		Email       string `json:"email"`
		Name        string `json:"name"`
		PhoneNumber string `json:"phone_number"`
	} `json:"customer"`
	Metadata          map[string]string `json:"metadata"`
	Reference         string            `json:"reference"`
	SavePaymentMethod bool              `json:"save_payment_method"`
}

type checkoutRes struct {
	CheckoutID        string         `json:"checkout_id"`
	CheckoutURL       string         `json:"checkout_url"`
	Status            CheckoutStatus `json:"status"`
	ExpiresAt         string         `json:"expires_at"`
	CreatedAt         string         `json:"created_at"`
	Reference         string         `json:"reference"`
	PlatformFee       string         `json:"platform_fee"`
	DestinationAmount string         `json:"destination_amount"`
}
