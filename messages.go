package agentphone

import(
	"context"
	"net/http"
)

// MessagesService handles the /messages endpoints: sending SMS/iMessage
// and reacting to messages.
type MessagesService struct {
	client *Client
}


// Message represents a single SMS/iMessage message.
type Message struct {
	ID 				string	`json:"id"`
	ConversationID	string	`json:"conversationId,omitempty"`
	FromNumber		string	`json:"fromNumber,omitempty"`
	ToNumber		string	`json:"toNumber,omitempty"`
	Body           	string 	`json:"body"`
	Direction      	string 	`json:"direction,omitempty"`

	// A message can be accepted (201) and still fail later at the
	// carrier; that failure lands here asynchronously, not as an API
	// error, so poll the message or use a webhook rather than expecting it
	// on the send response.
	Status 			string	`json:"status,omitempty"`
	FailureReason	string	`json:"failureReason,omitempty"`
	ReceivedAt 		string 	`json:"receivedAt,omitempty"`
}


// ListMessagesResponse is the response shape used by endpoints that return
// a page of messages (e.g. NumbersService.GetMessages). Messages use
// cursor-based pagination,
type ListMessagesResponse struct {
	Messages []Message `json:"data"`
	CursorPageInfo
}


// SendMessageParams are the parameters for sending a message. Exactly how
// the sender is identified (AgentID vs. NumberID)
type SendMessageParams struct {
	AgentID  string `json:"agentId,omitempty"`
	NumberID string `json:"numberId,omitempty"`
	ToNumber string `json:"toNumber"`
	Body     string `json:"body"`
}


// Send sends a new message.
func (s *MessagesService) send(ctx context.Context, params *SendMessageParams) (*Message, error) {
	var msg Message
	err := s.client.request(ctx, http.MethodPost, "/messages", params, &msg)
	return &msg, err
}


// Reaction represents a reaction (e.g. an emoji tapback) sent to a message.
type Reaction struct {
	ID        string `json:"id"`
	MessageID string `json:"messageId"`
	Emoji     string `json:"emoji"`
	CreatedAt string `json:"createdAt,omitempty"`
}


// SendReactionParams are the parameters for reacting to a message.
type SendReactionParams struct {
	Emoji string `json:"emoji"`
}


// SendReaction reacts to an existing message with an emoji.
func (s *MessagesService) SendReaction(ctx context.Context, messageID string, params *SendReactionParams) (*Reaction, error) {
	var reaction Reaction
	err := s.client.request(ctx, http.MethodPost, "/messages/"+messageID+"/reactions", params, &reaction)
	return &reaction, err
}
