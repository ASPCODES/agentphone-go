package agentphone

import(
	"encoding/json"
	"fmt"
)


// APIError represents an error response from the AgentPhone API. It's the base type for every error request() returns on a non-2xx response, and satisfies the standard error interface on its own.
type APIError struct {
	StatusCode		int
	Message 		string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("agentphone: %d %s", e.StatusCode, e.Message)
}


// AuthenticationError is returned when the API key is missing, invalid, or revoked (HTTP 401).
type AuthenticationError struct { *APIError }

// NotFoundError is returned when the requested resource doesn't exist (HTTP 404).
type NotFoundError struct { *APIError }

// RateLimitError is returned when the account has exceeded its request rate limit (HTTP 429).
type RateLimitError struct { *APIError }

// ValidationError is returned when the request body or parameters fail validation (HTTP 400 or 422).
type ValidationError struct { *APIError }


// Callers can check for a specific error with errors.As, e.g.:

// var notFound *agentphone.NotFoundError
//	if errors.As(err, &notFound) {
//	    // handle
//	}

type errorBody struct {
	Detail 		string	`json:"detail"`
	Message		string	`json:"message"`
	Error 		string	`json:"error"`
}


// parseAPIError converts a non-2xx HTTP response into a typed error.
func parseAPIError(statusCode int, body []byte) error {
	base := &APIError {
		StatusCode: statusCode,
		Message: 	extractErrorMessage(body),
	}

	switch statusCode {
	case 401:
		return &AuthenticationError{base}
	case 404:
		return &NotFoundError{base}
	case 429:
		return &RateLimitError{base}
	case 400, 422:
		return &ValidationError{base}
	default:
		return base
	}
}


// extractErrorMessage tries a few common JSON error shapes before falling back to the raw response body as a string. This keeps the SDK working even if AgentPhone's exact error format differs from what we expect, or changes later.
func extractErrorMessage(body []byte) string {
	var eb errorBody
	if err := json.Unmarshal(body, &eb); err == nil {
		switch {
		case eb.Detail != "":
			return eb.Detail
		case eb.Message != "":
			return eb.Message
		case eb.Error != "":
			return eb.Error
		}
	}

	if len(body) == 0 {
		return "unknown error"
	}

	return string(body)
}
