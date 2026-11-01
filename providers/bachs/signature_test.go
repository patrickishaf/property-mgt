package bachs

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/patrickishaf/property-mgt/config"
)

func sign(secret string, timestamp int64, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%d.%s", timestamp, body)))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestVerifySignature(t *testing.T) {
	const secret = "whsec_test"
	body := `{"type":"checkout.completed"}`
	now := time.Unix(1_800_000_000, 0)
	ts := now.Unix()
	good := sign(secret, ts, body)
	cases := []struct {
		name   string
		header string
		body   string
		secret string
		want   bool
	}{
		{"valid", fmt.Sprintf("t=%d,v1=%s", ts, good), body, secret, true},
		{"spaces after commas", fmt.Sprintf("t=%d, v1=%s", ts, good), body, secret, true},
		{"one of several v1 matches", fmt.Sprintf("t=%d,v1=deadbeef,v1=%s", ts, good), body, secret, true},
		{"wrong secret", fmt.Sprintf("t=%d,v1=%s", ts, good), body, "other", false},
		{"tampered body", fmt.Sprintf("t=%d,v1=%s", ts, good), body + " ", secret, false},
		{"too old", fmt.Sprintf("t=%d,v1=%s", ts-301, sign(secret, ts-301, body)), body, secret, false},
		{"too far ahead", fmt.Sprintf("t=%d,v1=%s", ts+301, sign(secret, ts+301, body)), body, secret, false},
		{"within tolerance", fmt.Sprintf("t=%d,v1=%s", ts-299, sign(secret, ts-299, body)), body, secret, true},
		{"timestamp not signed as given", fmt.Sprintf("t=%d,v1=%s", ts+1, good), body, secret, false},
		{"no timestamp", "v1=" + good, body, secret, false},
		{"bad timestamp", "t=soon,v1=" + good, body, secret, false},
		{"no signature", fmt.Sprintf("t=%d", ts), body, secret, false},
		{"empty header", "", body, secret, false},
		{"no secret configured", fmt.Sprintf("t=%d,v1=%s", ts, sign("", ts, body)), body, "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := VerifySignature(c.header, []byte(c.body), c.secret, SignatureTolerance, now); got != c.want {
				t.Fatalf("VerifySignature = %v, want %v", got, c.want)
			}
		})
	}
}

func TestVerifyWebhook(t *testing.T) {
	body := []byte(`{}`)
	signed := http.Header{}
	signed.Set(SignatureHeader, fmt.Sprintf("t=%d,v1=%s", time.Now().Unix(), sign("s3cret", time.Now().Unix(), "{}")))
	cases := []struct {
		name, env, secret string
		header            http.Header
		want              bool
	}{
		{"dev without secret accepts unsigned", "development", "", http.Header{}, true},
		{"production without secret rejects", "production", "", signed, false},
		{"secret requires a signature, even in dev", "development", "s3cret", http.Header{}, false},
		{"valid signature accepted", "production", "s3cret", signed, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			client := CreateClient(&config.BachsConfig{WebhookSecret: c.secret}, &config.AppConfig{Env: c.env})
			if got := client.VerifyWebhook(c.header, body); got != c.want {
				t.Fatalf("VerifyWebhook = %v, want %v", got, c.want)
			}
		})
	}
}
