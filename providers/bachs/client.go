package bachs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/patrickishaf/property-mgt/config"
)

type Client struct {
	appCfg *config.AppConfig
	cfg    *config.BachsConfig
	http   *http.Client
}

func CreateClient(cfg *config.BachsConfig, appCfg *config.AppConfig) *Client {
	return &Client{
		appCfg: appCfg,
		cfg:    cfg,
		http:   &http.Client{Timeout: 20 * time.Second},
	}
}

// Stubbed reports whether checkouts are faked locally: in development when no
// API key is configured. Stub checkouts are completed with the development-only
// POST /payments/checkout/{reference}/complete route.
func (this *Client) Stubbed() bool {
	return this.cfg.APIKey == "" && this.appCfg.Env == "development"
}

// InitiateCheckout creates a hosted checkout session for the payment.
func (this *Client) InitiateCheckout(dto CreateCheckoutSessionBody) (*CheckoutRes, error) {
	if this.Stubbed() {
		return this.stubCheckout(dto), nil
	}
	if this.cfg.APIKey == "" || this.cfg.APIBaseURL == "" {
		return nil, fmt.Errorf("bachs: BACHS_API_BASE_URL and BACHS_API_KEY must be set")
	}
	raw, err := json.Marshal(dto)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, this.cfg.APIBaseURL+"/v1/checkout-sessions", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+this.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := this.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bachs: create checkout: %w", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("bachs: read checkout response: %w", err)
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, fmt.Errorf("bachs: create checkout: status %d: %s", res.StatusCode, body)
	}
	var checkout CheckoutRes
	if err := json.Unmarshal(body, &checkout); err != nil {
		return nil, fmt.Errorf("bachs: decode checkout response: %w", err)
	}
	if checkout.CheckoutURL == "" {
		return nil, fmt.Errorf("bachs: checkout response has no checkout_url")
	}
	return &checkout, nil
}

func (this *Client) stubCheckout(dto CreateCheckoutSessionBody) *CheckoutRes {
	now := time.Now()
	expiresAt := now.Add(time.Duration(this.appCfg.CheckoutExpiryMinutes) * time.Minute)
	return &CheckoutRes{
		CheckoutID:  "stub_" + dto.Reference,
		CheckoutURL: this.appCfg.WebAppURL + "/checkout/stub?reference=" + url.QueryEscape(dto.Reference),
		Status:      OpenCheckoutStatus,
		ExpiresAt:   expiresAt.Format(time.RFC3339),
		CreatedAt:   now.Format(time.RFC3339),
		Reference:   dto.Reference,
	}
}

// GetCheckout fetches a checkout session.
//
// TODO: call the Bachs API once its retrieve endpoint is known.
func (this *Client) GetCheckout(checkoutID string) (*CheckoutRes, error) {
	return nil, ErrNotImplemented
}
