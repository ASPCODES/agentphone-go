package agentphone

import(
	"context"
	"net/url"
	"net/http"
)


// NumbersService handles the /numbers endpoints: buying, listing, and
// managing phone numbers.
type NumbersService struct {
	client *Client
}

// Number represents a phone number owned on the account.
type Number struct {
	ID 			string		`json:"id"`
	PhoneNumber	string		`json:"phoneNumber"`
	Type 		string		`json:"type"`
	AgentID		string		`json:"agentId,omitempty"`
	CreatedAt	string		`json:"createdAt,omitempty"`
}


// ListNumbersResponse is the response from ListNumbers. Numbers uses offset-based pagination.
type ListNumbersResponse struct {
	Numbers []Number `json:"data"`
	OffsetPageInfo
}


// List returns the numbers on the account. params may be nil.
func (s *NumbersService) List(ctx context.Context, params *ListParams) (*ListNumbersResponse, error) {
	var resp ListNumbersResponse
	err := s.client.request(ctx, http.MethodGet, "/numbers"+params.toQuery(), nil, &resp)
	return &resp, err
}


// CreateNumberParams are the parameters for buying/creating a number.
type CreateNumberParams struct {
	PhoneNumber		string	`json:"phoneNumber,omitempty"`
	AreaCode		string	`json:"areaCode,omitempty"`
	Type			string	`json:"type,omitempty"`
}


// Create purchases/provisions a new number.
func (s *NumbersService) Create(ctx context.Context, params *CreateNumberParams) (*Number, error) {
	var num Number
	err := s.client.request(ctx, http.MethodGet, "/numbers", params, &num)
	return &num, err
}


// AvailableNumber is a number that can be purchased.
type AvailableNumber struct {
	PhoneNumber		string	`json:"phoneNumber"`
	Type 			string	`json:"type"`
	Price			string	`json:"price"`
}


// ListAvailableNumbersResponse is the response from ListAvailable.
type ListAvailableNumbersResponse struct {
	Numbers []AvailableNumber  `json:"data"`
	OffsetPageInfo
}

// ListAvailableNumbersParams filters the search for purchasable numbers.
type ListAvailableNumbersParams struct {
	AreaCode	string
	Country		string
	Type		string
}

func (p *ListAvailableNumbersParams) toQuery() string {
	if p == nil {
		return ""
	}
	
	q := url.Values{}
	if p.AreaCode != "" {
		q.Set("areaCode", p.AreaCode)
	}
	if p.Country != "" {
		q.Set("country", p.Country)
	}
	if p.Type != "" {
		q.Set("type", p.Type)
	}
	if encoded := q.Encode(); encoded != "" {
		return "?" + encoded
	}
	
	return ""
}

// ListAvailable searches for numbers available for purchase. params may be nil.
func (s *NumbersService) ListAvailable(ctx context.Context, params *ListAvailableNumbersParams) (*ListAvailableNumbersResponse, error) {
	var resp ListAvailableNumbersResponse
	err := s.client.request(ctx, http.MethodGet, "/numbers/available"+params.toQuery(), nil, &resp)
	return &resp, err
}


// LookupResult is the result of a number lookup (carrier/line-type info)
type LookupResult struct {
	PhoneNumber		string	`json:"phonenumber"`
	Valid 			string	`json:"valid,omitempty"`
	Carrier 		string	`json:"carrier,omitempty"`
	LineType    	string 	`json:"lineType,omitempty"`
}


// Lookup checks carrier and line-type info for a phone number.Screening a number this way before sending is cheaper than paying for a message that fails delivery
func (s *NumbersService) Lookup(ctx context.Context, phoneNumber string) (*LookupResult, error) {
	q := url.Values{}
	q.Set("phoneNumber", phoneNumber)

	var result LookupResult
	err := s.client.request(ctx, http.MethodGet, "/numbers/lookup?"+q.Encode(), nil, &result)
	return &result, err
}


// Get retrieves a single number by ID.
func (s *NumbersService) Get(ctx context.Context, numberID string) (*Number, error) {
	var num Number
	err := s.client.request(ctx, http.MethodGet, "/numbers/"+numberID, nil, &num)
	return &num, err
}


// Delete a number.
func (s *NumbersService) Delete(ctx context.Context, numberID string) error {
	return s.client.request(ctx, http.MethodGet, "/numbers/"+numberID, nil, nil)
}


func (s *NumbersService) GetMessages(ctx context.Context, numberID string, params *CursorListParams) (*ListMessagesResponse, error) {
	var resp ListMessagesResponse
	err := s.client.request(ctx, http.MethodGet, "/numbers/"+numberID+"/messages"+params.toQuery(), nil, &resp)
	return &resp, err
}
