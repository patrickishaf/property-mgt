package bachs

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// SignatureHeader carries a webhook's signature as "t=<unix seconds>,v1=<hex>",
// possibly with several v1 entries while Bachs rotates secrets. The signature
// covers the raw request body, so it must be checked before the JSON is parsed.
const SignatureHeader = "X-Bachs-Signature-V2"

// SignatureTolerance is how far a webhook's timestamp may be from now before
// it is rejected as a replay.
const SignatureTolerance = 5 * time.Minute

// VerifySignature reports whether header signs rawBody with secret: one of its
// v1 values must equal hex(HMAC-SHA256(secret, "<t>." + rawBody)), and t must
// be within tolerance of now. Malformed headers fail verification.
func VerifySignature(header string, rawBody []byte, secret string, tolerance time.Duration, now time.Time) bool {
	if secret == "" {
		return false
	}
	var timestamp int64
	haveTimestamp := false
	var signatures []string
	for _, part := range strings.Split(header, ",") {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		switch key {
		case "t":
			t, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return false
			}
			timestamp, haveTimestamp = t, true
		case "v1":
			signatures = append(signatures, value)
		}
	}
	if !haveTimestamp || len(signatures) == 0 {
		return false
	}
	if age := now.Sub(time.Unix(timestamp, 0)); age > tolerance || age < -tolerance {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strconv.FormatInt(timestamp, 10) + "."))
	mac.Write(rawBody)
	expected := hex.EncodeToString(mac.Sum(nil))
	for _, signature := range signatures {
		if hmac.Equal([]byte(expected), []byte(signature)) {
			return true
		}
	}
	return false
}

// VerifyWebhook reports whether a webhook delivery came from Bachs. With no
// BACHS_WH_SECRET in development, unsigned deliveries are accepted so webhooks
// can be simulated locally; otherwise a valid signature is required.
func (this *Client) VerifyWebhook(header http.Header, rawBody []byte) bool {
	if this.cfg.WebhookSecret == "" {
		return this.appCfg.Env == "development"
	}
	return VerifySignature(header.Get(SignatureHeader), rawBody, this.cfg.WebhookSecret, SignatureTolerance, time.Now())
}
