package agentphone

import(
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)


const(
	defaultBaseURL = "https://api.agentphone.ai/v1"
	defaultTimeOut = 30 * time.Second
)

// Client is the AgentPhone API client. Create one with NewClient.
type Client struct {
	apiKey		string
	baseURL		string
	httpClient	*http.Client
	
	Agents			*AgentsService
	Numbers			*NumbersService
	Calls			*CallsService
	Messages		*MessagesService
	Conversations	*ConversationsService
	Contacts		*ContactsService
	ContactCards	*ContactCardsService
	Webhooks      	*WebhooksService
	Verification  	*VerificationService
	Usage			*UsageService
	SubAccounts		*SubAccountsService
	SIPTrunks		*SIPTrunksService
	WhatsApp		*WhatsAppService
	Registration  	*RegistrationService
	Location      	*LocationService
}

// clientOptions collects everything the Option functions configure, before
// a Client is actually built.
type clientOptions struct {
	baseURL			string
	httpClient		*http.Client
	timeout 		*time.Duration		
}


// Option configures optional Client behavior. Pass zero or more Options to
// NewClient. Options can be passed in any order.
type Option func(*clientOptions)

func WithBaseURL(baseURL string) Option {
	return func(o *clientOptions) {
		o.baseURL = baseURL
	}
}


// WithHTTPClient overrides the default *http.Client used to send requests.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(o *clientOptions) {
		o.httpClient = httpClient
	}
}


// WithTimeout overrides the request timeout.Applies regardless of whether it's passed before or after WithHTTPClient.
func WithTimeout(timeout time.Duration) Option {
	return func(o *clientOptions) {
		o.timeout = &timeout
	}
}



// NewClient creates a new AgentPhone API client. apiKey is required; every other setting has a sensible default and can be overridden with Option functions, in any order.
func NewClient(apiKey string, opts ...Option) *Client {
	cfg := &clientOptions{baseURL: defaultBaseURL}

	for _, opt := range opts {
		opt(cfg)
	}

	httpClient := cfg.httpClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeOut}
	}
	if cfg.timeout == nil {
		httpClient.Timeout = *cfg.timeout
	}

	c := &Client {
		apiKey: 		apiKey,
		baseURL: 		strings.TrimRight(cfg.baseURL, "/"),
		httpClient: 	httpClient,
	}
}