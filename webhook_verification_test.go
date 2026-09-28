package agentphone

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestVerifyWebhook_AcceptsValidSignature(t *testing.T) {
	payload := []byte(`{"event":"message.received"}`)
	timestamp := time.Now().Unix()
	timestampString := strconvFormatInt(timestamp)
	signature := documentedWebhookSignature("whsec_test", timestampString, payload)

	if err := VerifyWebhook(payload, signature, "whsec_test", timestampString, time.Minute); err != nil {
		t.Fatalf("VerifyWebhook() error: %v", err)
	}
}

func TestVerifyWebhook_RejectsBadSignatureAndStaleTimestamp(t *testing.T) {
	payload := []byte(`{"event":"message.received"}`)
	timestampString := strconvFormatInt(time.Now().Unix())
	err := VerifyWebhook(payload, "invalid", "whsec_test", timestampString, time.Minute)
	var verifyErr *WebhookVerificationError
	if !errors.As(err, &verifyErr) || !strings.Contains(verifyErr.Message, "signature mismatch") {
		t.Fatalf("error = %T %v, want signature mismatch verification error", err, err)
	}

	stale := strconvFormatInt(time.Now().Add(-10 * time.Minute).Unix())
	err = VerifyWebhook(payload, documentedWebhookSignature("whsec_test", stale, payload), "whsec_test", stale, time.Minute)
	if !errors.As(err, &verifyErr) || !strings.Contains(verifyErr.Message, "outside the allowed tolerance") {
		t.Fatalf("error = %T %v, want stale timestamp verification error", err, err)
	}
}

func TestVerifyWebhook_RejectsInvalidTimestampAndSupportsDisabledTolerance(t *testing.T) {
	payload := []byte(`payload`)
	err := VerifyWebhook(payload, "signature", "secret", "not-a-timestamp", time.Minute)
	var verifyErr *WebhookVerificationError
	if !errors.As(err, &verifyErr) || verifyErr.Message != "invalid or missing timestamp" {
		t.Fatalf("error = %T %v, want invalid timestamp verification error", err, err)
	}

	signature := documentedWebhookSignature("secret", "", payload)
	if err := VerifyWebhook(payload, signature, "secret", "", -1); err != nil {
		t.Fatalf("VerifyWebhook() with disabled tolerance error: %v", err)
	}
}

func TestVerifyWebhook_UsesDocumentedSignatureFormat(t *testing.T) {
	payload := []byte(`{"event":"agent.message"}`)
	timestamp := strconvFormatInt(time.Now().Unix())
	signature := documentedWebhookSignature("whsec_test", timestamp, payload)
	if err := VerifyWebhook(payload, signature, "whsec_test", timestamp, time.Minute); err != nil {
		t.Fatalf("VerifyWebhook() rejected documented signature: %v", err)
	}
	if err := VerifyWebhook(payload, strings.TrimPrefix(signature, "sha256="), "whsec_test", timestamp, time.Minute); err == nil {
		t.Fatal("VerifyWebhook() accepted a signature without the documented sha256= prefix")
	}
}

func TestConstructEvent_VerifiesThenParsesPayload(t *testing.T) {
	payload := []byte(`{"event":"agent.message","channel":"sms","timestamp":"2025-01-15T12:00:00Z","agentId":"agt_1","data":{"conversationId":"conv_1","numberId":"num_1","message":"Hi","mediaUrl":"https://example.com/photo.jpg","from":"+14155550100","to":"+14155550101","direction":"inbound","receivedAt":"2025-01-15T12:00:00Z"},"recentHistory":[{"direction":"inbound","content":"Hello","channel":"sms","at":"2025-01-15T11:59:00Z"}],"conversationState":{"intent":"support"}}`)
	timestamp := strconvFormatInt(time.Now().Unix())
	signature := documentedWebhookSignature("whsec_test", timestamp, payload)

	event, err := ConstructEvent(payload, signature, "whsec_test", timestamp, time.Minute)
	if err != nil {
		t.Fatalf("ConstructEvent() error: %v", err)
	}
	if event.Event != "agent.message" || event.Channel != "sms" || event.AgentID != "agt_1" || event.Timestamp != "2025-01-15T12:00:00Z" || event.Data.Message != "Hi" || event.Data.MediaURL != "https://example.com/photo.jpg" || event.Data.From != "+14155550100" || event.Data.To != "+14155550101" || event.Data.ConversationID != "conv_1" || event.Data.NumberID != "num_1" || len(event.RecentHistory) != 1 || event.RecentHistory[0].Channel != "sms" || event.ConversationState["intent"] != "support" {
		t.Errorf("event = %+v", event)
	}

	if _, err := ConstructEvent(payload, "bad-signature", "whsec_test", timestamp, time.Minute); err == nil {
		t.Fatal("ConstructEvent() accepted an invalid signature")
	}
	invalidJSON := []byte("not-json")
	invalidSignature := documentedWebhookSignature("whsec_test", timestamp, invalidJSON)
	if _, err := ConstructEvent(invalidJSON, invalidSignature, "whsec_test", timestamp, time.Minute); err == nil || !strings.Contains(err.Error(), "decoding webhook event") {
		t.Fatalf("error = %v, want verified JSON decode error", err)
	}
}

func TestConstructEvent_PreservesCallEndedMetadata(t *testing.T) {
	payload := []byte(`{"event":"agent.call_ended","channel":"voice","timestamp":"2025-01-15T14:05:30Z","agentId":"agt_abc123","data":{"callId":"call_ghi012","numberId":"num_xyz789","from":"+15559876543","to":"+15551234567","direction":"inbound","status":"completed","startedAt":"2025-01-15T14:00:00Z","endedAt":"2025-01-15T14:05:30Z","durationSeconds":330,"disconnectionReason":"agent_hangup","transcript":[{"role":"agent","content":"Hello"},{"role":"user","content":"I need help"}],"summary":"Customer called about an order.","userSentiment":"Positive","callSuccessful":true}}`)
	timestamp := strconvFormatInt(time.Now().Unix())
	signature := documentedWebhookSignature("whsec_test", timestamp, payload)

	event, err := ConstructEvent(payload, signature, "whsec_test", timestamp, time.Minute)
	if err != nil {
		t.Fatalf("ConstructEvent() error: %v", err)
	}
	if event.AgentID != "agt_abc123" || event.Data.CallID != "call_ghi012" || event.Data.NumberID != "num_xyz789" || event.Data.From != "+15559876543" || event.Data.To != "+15551234567" || event.Data.DurationSeconds != 330 || event.Data.DisconnectionReason != "agent_hangup" || event.Data.Summary != "Customer called about an order." || event.Data.UserSentiment != "Positive" || event.Data.CallSuccessful == nil || !*event.Data.CallSuccessful {
		t.Errorf("event metadata = %+v", event)
	}
	if len(event.Data.TranscriptTurns) != 2 || event.Data.TranscriptTurns[1].Content != "I need help" {
		t.Errorf("transcript turns = %+v", event.Data.TranscriptTurns)
	}
}

func documentedWebhookSignature(secret, timestamp string, payload []byte) string {
	signedPayload := make([]byte, 0, len(timestamp)+1+len(payload))
	signedPayload = append(signedPayload, timestamp...)
	signedPayload = append(signedPayload, '.')
	signedPayload = append(signedPayload, payload...)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(signedPayload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func strconvFormatInt(value int64) string {
	return strconv.FormatInt(value, 10)
}
