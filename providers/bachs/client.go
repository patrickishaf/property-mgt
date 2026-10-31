package bachs

import (
	"net/url"
	"time"

	"github.com/patrickishaf/property-mgt/config"
)

type Client struct {
	appCfg *config.AppConfig
	cfg    *config.BachsConfig
}

func CreateClient(cfg *config.BachsConfig, appCfg *config.AppConfig) *Client {
	return &Client{
		appCfg: appCfg,
		cfg:    cfg,
	}
}

// InitiateCheckout creates a hosted checkout session for the payment.
//
// TODO: call the Bachs checkout API once its endpoint and auth details are known.
// Until then this returns a local stub checkout that is completed with the
// development-only POST /payments/checkout/{reference}/complete route.
func (this *Client) InitiateCheckout(dto CreateCheckoutSessionBody) (*CheckoutRes, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(this.appCfg.CheckoutExpiryMinutes) * time.Minute)
	return &CheckoutRes{
		CheckoutID:  "stub_" + dto.Reference,
		CheckoutURL: this.appCfg.WebAppURL + "/checkout/stub?reference=" + url.QueryEscape(dto.Reference),
		Status:      OpenCheckoutStatus,
		ExpiresAt:   expiresAt.Format(time.RFC3339),
		CreatedAt:   now.Format(time.RFC3339),
		Reference:   dto.Reference,
	}, nil
}

// GetCheckout fetches a checkout session.
//
// TODO: call the Bachs API; not needed while checkouts are stubbed.
func (this *Client) GetCheckout(checkoutID string) (*CheckoutRes, error) {
	return nil, ErrNotImplemented
}
