package messaging

import (
	"testing"

	"github.com/patrickishaf/property-mgt/internal/db/model"
)

func TestPersonalise(t *testing.T) {
	got := personalise("Dear {name}, {unit} at {property}. Thanks {name}. Ref {code}.", "Ada Obi", "1A", "Palm Court")
	want := "Dear Ada Obi, 1A at Palm Court. Thanks Ada Obi. Ref {code}."
	if got != want {
		t.Fatalf("personalise() = %q, want %q", got, want)
	}
	if got := personalise("No placeholders", "x", "y", "z"); got != "No placeholders" {
		t.Fatalf("expected text without placeholders to be unchanged, got %q", got)
	}
}

func TestTemplatesUseOnlyKnownPlaceholders(t *testing.T) {
	for _, tmpl := range templates {
		filled := personalise(tmpl.Body+tmpl.Subject, "n", "u", "p")
		for i := range filled {
			if filled[i] == '{' {
				t.Fatalf("template %q has an unknown placeholder: %q", tmpl.Key, filled)
			}
		}
	}
}

func TestCheckSendRequest(t *testing.T) {
	recipient := func(id uint, email string, phone string) recipientReq {
		return recipientReq{TenantID: id, FullName: "Ada Obi", Email: email, PhoneNumber: phone}
	}
	cases := []struct {
		name        string
		dto         sendMessageReq
		wantCode    int
		wantSubject string
	}{
		{"in-app needs no contact details", sendMessageReq{Channel: model.MessageChannelInApp, Body: "Hi", Recipients: []recipientReq{recipient(1, "", "")}}, 0, ""},
		{"sms without phone", sendMessageReq{Channel: model.MessageChannelSMS, Body: "Hi", Recipients: []recipientReq{recipient(1, "a@b.c", "")}}, 400, ""},
		{"sms with phone", sendMessageReq{Channel: model.MessageChannelSMS, Body: "Hi", Recipients: []recipientReq{recipient(1, "", "080")}}, 0, ""},
		{"email without email", sendMessageReq{Channel: model.MessageChannelEmail, Subject: "S", Body: "Hi", Recipients: []recipientReq{recipient(1, "", "080")}}, 400, ""},
		{"email without subject or template", sendMessageReq{Channel: model.MessageChannelEmail, Body: "Hi", Recipients: []recipientReq{recipient(1, "a@b.c", "")}}, 400, ""},
		{"email uses template subject", sendMessageReq{Channel: model.MessageChannelEmail, Template: "payment_reminder", Body: "Hi", Recipients: []recipientReq{recipient(1, "a@b.c", "")}}, 0, "Rent payment reminder"},
		{"subject overrides template", sendMessageReq{Channel: model.MessageChannelEmail, Template: "payment_reminder", Subject: "Custom", Body: "Hi", Recipients: []recipientReq{recipient(1, "a@b.c", "")}}, 0, "Custom"},
		{"unknown template", sendMessageReq{Channel: model.MessageChannelInApp, Template: "nope", Body: "Hi", Recipients: []recipientReq{recipient(1, "", "")}}, 400, ""},
		{"blank body", sendMessageReq{Channel: model.MessageChannelInApp, Body: "   ", Recipients: []recipientReq{recipient(1, "", "")}}, 400, ""},
		{"duplicate recipients", sendMessageReq{Channel: model.MessageChannelInApp, Body: "Hi", Recipients: []recipientReq{recipient(1, "", ""), recipient(1, "", "")}}, 400, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			subject, res := checkSendRequest(tc.dto)
			got := 0
			if res != nil {
				got = res.StatusCode
			}
			if got != tc.wantCode {
				t.Fatalf("status = %d, want %d (%+v)", got, tc.wantCode, res)
			}
			if res == nil && subject != tc.wantSubject {
				t.Fatalf("subject = %q, want %q", subject, tc.wantSubject)
			}
		})
	}
}
