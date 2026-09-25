package agentphone

import(
	"context"
	"net/http"
)


// AgentsService handles the /agents endpoints: creating and configuring AI agent personas, and their attached numbers, conversations, and calls.
type AgentsService struct {
	client  *Client
}


// Agent represents an AI agent persona (voice mode, prompt, voice settings) that can have phone numbers attached to it.
type Agent struct {
	ID 						string	`json:"id"`
	Name					string	`json:"name"`
	Description 			string 	`json:"description,omitempty"`
	
	// VoiceMode is "webhook" (default — forwards call transcripts to your webhook) built-in LLM handles the call using SystemPrompt.
	VoiceMode   			string 	`json:"voiceMode,omitempty"`
	SystemPrompt    		string 	`json:"systemPrompt,omitempty"`
	BeginMessage    		string 	`json:"beginMessage,omitempty"`
	Voice 					string	`json:"voice,omitempty"`
	ModelTier       		string 	`json:"modelTier,omitempty"`
	STTMode         		string 	`json:"sttMode,omitempty"`
	AmbientSound    		string 	`json:"ambientSound,omitempty"`
	DenoisingMode   		string 	`json:"denoisingMode,omitempty"`
	TransferNumber  		string 	`json:"transferNumber,omitempty"`
	VoicemailMessage 		string 	`json:"voicemailMessage,omitempty"`
	Language				string	`json:"language,omitempty"`
	VoiceSpeed              float64 `json:"voiceSpeed,omitempty"`
	InterruptionSensitivity float64 `json:"interruptionSensitivity,omitempty"`
	EnableBackchannel       bool    `json:"enableBackchannel,omitempty"`
	MaxSilenceMs            int     `json:"maxSilenceMs,omitempty"`
	EnableMessaging         bool    `json:"enableMessaging,omitempty"`
 
	CreatedAt 				string  `json:"createdAt,omitempty"`
	Numbers   []Number `json:"numbers,omitempty"`
}

// CreateAgentParams are the parameters for creating an agent. Name is required; everything else is optional and uses the platform default when omitted. VoiceSpeed, InterruptionSensitivity, EnableBackchannel, MaxSilenceMs and EnableMessaging are pointers because their zero value (0 / false) is meaningful and distinct from "unset"

