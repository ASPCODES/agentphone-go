package agentphone

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// This file verifies INCOMING webhook deliveries (the requests AgentPhone
// sends to your server) — a different concern from WebhooksService in
// webhooks.go, which manages your webhook *configuration* via the API.

// WebhookVerificationError indicates a webhook delivery's signature is
// invalid, or its timestamp fell outside the allowed replay window.
type WebhookVerificationError struct {
	Message string
}

func (e *WebhookVerificationError) Error() string {
	return "agentphone: webhook verification failed: " + e.Message
}

// DefaultWebhookTolerance is the default max age (5 minutes) allowed between a delivery's timestamp and now
const DefaultWebhookTolerance = 5 * time.Minute

// VerifyWebhook checks a webhook delivery's signature and, unless disabled, that its timestamp is recent enough to rule out a replayed request.
func VerifyWebhook(payload []byte, signature, secret, timestamp string, tolerance time.Duration) error {
	if tolerance == 0 {
		tolerance = DefaultWebhookTolerance
	}

	if tolerance > 0 {
		ts, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			return &WebhookVerificationError{Message: "invalid or missing timestamp"}
		}
		age := time.Since(time.Unix(ts, 0))
		if age < 0 {
			age = -age
		}
		if age > tolerance {
			return &WebhookVerificationError{Message: "timestamp outside the allowed tolerance (possible replay)"}
		}
	}

	expected := signWebhookPayload(secret, timestamp, payload)
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return &WebhookVerificationError{Message: "signature mismatch"}
	}
	return nil
}

// signWebhookPayload computes the expected signature for a delivery.
func signWebhookPayload(secret, timestamp string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("."))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// WebhookHistoryItem is one turn of recent conversation context included on a webhook event, for use as LLM context.
type WebhookHistoryItem struct {
	Direction        string `json:"direction"` // "inbound" or "outbound".
	Content          string `json:"content"`
	Channel          string `json:"channel,omitempty"`
	At               string `json:"at,omitempty"`
	SenderIdentifier string `json:"senderIdentifier,omitempty"`
}

// WebhookTranscriptTurn is one turn in an agent.call_ended transcript.
type WebhookTranscriptTurn struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// WebhookReplyTo describes the message an inbound WhatsApp reply refers to.
type WebhookReplyTo struct {
	MessageID string   `json:"messageId,omitempty"`
	Message   string   `json:"message,omitempty"`
	MediaURLs []string `json:"mediaUrls,omitempty"`
}

// WebhookGroupParticipant is a member of an iMessage group conversation.
type WebhookGroupParticipant struct {
	Identifier string  `json:"identifier"`
	Name       *string `json:"name,omitempty"`
}

// WebhookGroup contains group-chat metadata included with iMessage events.
type WebhookGroup struct {
	IsGroup      bool                      `json:"isGroup"`
	GroupID      string                    `json:"groupId,omitempty"`
	GroupName    *string                   `json:"groupName,omitempty"`
	GroupIconURL *string                   `json:"groupIconUrl,omitempty"`
	Participants []WebhookGroupParticipant `json:"participants,omitempty"`
}

// WebhookEventData carries the event-specific payload. Which fields are set depends on Channel: voice events carry Transcript
type WebhookEventData struct {
	ConversationID      string                  `json:"conversationId,omitempty"`
	NumberID            string                  `json:"numberId,omitempty"`
	CallID              string                  `json:"callId,omitempty"`
	Message             string                  `json:"message,omitempty"`
	MediaURL            string                  `json:"mediaUrl,omitempty"`
	MediaURLs           []string                `json:"mediaUrls,omitempty"`
	Direction           string                  `json:"direction,omitempty"`
	ReceivedAt          string                  `json:"receivedAt,omitempty"`
	Transcript          string                  `json:"transcript,omitempty"`
	TranscriptTurns     []WebhookTranscriptTurn `json:"-"`
	Confidence          float64                 `json:"confidence,omitempty"`
	Status              string                  `json:"status,omitempty"`
	From                string                  `json:"from,omitempty"`
	To                  string                  `json:"to,omitempty"`
	FromNumber          string                  `json:"fromNumber,omitempty"`
	StartedAt           string                  `json:"startedAt,omitempty"`
	EndedAt             string                  `json:"endedAt,omitempty"`
	DurationSeconds     int                     `json:"durationSeconds,omitempty"`
	DisconnectionReason string                  `json:"disconnectionReason,omitempty"`
	Summary             string                  `json:"summary,omitempty"`
	UserSentiment       string                  `json:"userSentiment,omitempty"`
	CallSuccessful      *bool                   `json:"callSuccessful,omitempty"`
	ReplyTo             *WebhookReplyTo         `json:"replyTo,omitempty"`
	Group               *WebhookGroup           `json:"group,omitempty"`
	SenderIdentifier    string                  `json:"senderIdentifier,omitempty"`
	ReactionType        string                  `json:"reactionType,omitempty"`
	MessageID           string                  `json:"messageId,omitempty"`
	MessageBody         string                  `json:"messageBody,omitempty"`
	MessageMediaURL     string                  `json:"messageMediaUrl,omitempty"`
	CreatedAt           string                  `json:"createdAt,omitempty"`
}

func (d *WebhookEventData) UnmarshalJSON(data []byte) error {
	type eventDataAlias WebhookEventData
	var decoded struct {
		*eventDataAlias
		Transcript json.RawMessage `json:"transcript"`
	}
	decoded.eventDataAlias = (*eventDataAlias)(d)
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	if len(decoded.Transcript) == 0 || bytes.Equal(decoded.Transcript, []byte("null")) {
		return nil
	}
	if decoded.Transcript[0] == '"' {
		return json.Unmarshal(decoded.Transcript, &d.Transcript)
	}
	return json.Unmarshal(decoded.Transcript, &d.TranscriptTurns)
}

// WebhookEvent is an incoming webhook delivery, parsed by ConstructEvent.
type WebhookEvent struct {
	Event             string                 `json:"event"`
	Channel           string                 `json:"channel"`
	Timestamp         string                 `json:"timestamp,omitempty"`
	AgentID           string                 `json:"agentId,omitempty"`
	Data              WebhookEventData       `json:"data"`
	RecentHistory     []WebhookHistoryItem   `json:"recentHistory,omitempty"`
	ConversationState map[string]interface{} `json:"conversationState,omitempty"`
}

// ConstructEvent verifies a webhook delivery (see VerifyWebhook) and, only if verification succeeds, parses its body into a WebhookEvent. Use this instead of decoding the body yourself so you never process an unverified payload.
func ConstructEvent(payload []byte, signature, secret, timestamp string, tolerance time.Duration) (*WebhookEvent, error) {
	if err := VerifyWebhook(payload, signature, secret, timestamp, tolerance); err != nil {
		return nil, err
	}

	var event WebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("agentphone: decoding webhook event: %w", err)
	}
	return &event, nil
}
