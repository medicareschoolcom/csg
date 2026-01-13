// Package medsup provides a client for the CSG Medicare Supplement API.
package medsup

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const defaultBaseURL = "https://api.csgactuarial.com/v1/med_supp"

// Gender represents the gender parameter for quotes.
type Gender string

const (
	Male   Gender = "M"
	Female Gender = "F"
)

// Boolean represents boolean query parameters as strings ("1" or "0").
type Boolean string

const (
	True  Boolean = "1"
	False Boolean = "0"
)

// Plan represents Medicare Supplement plan types.
type Plan string

const (
	PlanF Plan = "F"
	PlanG Plan = "G"
	PlanN Plan = "N"
)

// Quote represents a Medicare Supplement insurance quote.
type Quote struct {
	Age         int       `json:"age"`
	Gender      string    `json:"gender"`
	Plan        string    `json:"plan"`
	Tobacco     bool      `json:"tobacco"`
	Rate        QuoteRate `json:"rate"`
	CompanyBase Company   `json:"company_base"`
}

// QuoteRate contains premium rates for different payment frequencies.
type QuoteRate struct {
	Quarter    int `json:"quarter"`
	Annual     int `json:"annual"`
	SemiAnnual int `json:"semi_annual"`
	Month      int `json:"month"`
}

// Company represents an insurance company.
type Company struct {
	NameFull string `json:"name_full"`
	Name     string `json:"name"`
	NAIC     string `json:"naic"`
}

// TokenProvider provides authentication tokens for API requests.
type TokenProvider interface {
	GetToken(ctx context.Context) (string, error)
}

// Client is an HTTP client for CSG Medicare Supplement quotes API.
type Client struct {
	httpClient    *http.Client
	baseURL       string
	tokenProvider TokenProvider
	token         string
	mu            sync.RWMutex
}

// Option is a function that configures a Client.
type Option = func(*Client)

// NewClient creates a new Medicare Supplement quotes client.
func NewClient(httpClient *http.Client, tokenProvider TokenProvider, options ...Option) *Client {
	c := &Client{
		httpClient:    httpClient,
		tokenProvider: tokenProvider,
		baseURL:       defaultBaseURL,
	}

	for _, option := range options {
		option(c)
	}

	return c
}

// WithBaseURL configures the client to use a custom base URL.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// ListQuotesParams contains parameters for Medicare Supplement quote requests.
type ListQuotesParams struct {
	Zip5          string // "15963"
	Counties      []string
	Age           int     // 65
	Gender        Gender  // "M" or "F"
	Tobacco       Boolean // 0 or 1
	Select        *Boolean
	NAICS         []string
	Plan          Plan // "G"
	EffectiveDate *time.Time
}

// ListQuotes retrieves Medicare Supplement quotes from CSG API.
func (c *Client) ListQuotes(ctx context.Context, params ListQuotesParams) ([]*Quote, error) {
	// Build URL with query parameters
	url := c.baseURL + "/quotes.json"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set context
	req = req.WithContext(ctx)

	// Build query string
	q := req.URL.Query()

	// Zip5 required.
	q.Add("zip5", params.Zip5)

	// County optional (repeatable).
	for _, county := range params.Counties {
		q.Add("county", county)
	}

	// Age required.
	q.Add("age", strconv.Itoa(params.Age))

	// Gender required.
	q.Add("gender", string(params.Gender))

	// Tobacco required.
	q.Add("tobacco", string(params.Tobacco))

	// Select optional.
	if params.Select != nil {
		q.Add("select", string(*params.Select))
	}

	// NAIC optional (repeatable).
	for _, naic := range params.NAICS {
		q.Add("naic", naic)
	}

	// Plan required.
	q.Add("plan", string(params.Plan))

	// Effective date optional. Defaults to todays date.
	if params.EffectiveDate != nil {
		q.Add("effective_date", params.EffectiveDate.Format(time.DateOnly))
	}

	// Field
	q.Add("field", "age")
	q.Add("field", "gender")
	q.Add("field", "plan")
	q.Add("field", "tobacco")
	q.Add("field", "rate.quarter")
	q.Add("field", "rate.annual")
	q.Add("field", "rate.semi_annual")
	q.Add("field", "rate.month")
	q.Add("field", "company_base.name_full")
	q.Add("field", "company_base.name")
	q.Add("field", "company_base.naic")

	req.URL.RawQuery = q.Encode()

	// Execute authenticated request with automatic retry on 403
	res, err := c.doWithToken(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Read response body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response, status code: %d, body: %s", res.StatusCode, string(body))
	}

	quotes := []*Quote{}
	err = json.Unmarshal(body, &quotes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse json response: %w", err)
	}

	return quotes, nil
}

// ListCompanies retrieves the list of available insurance companies.
// This endpoint does not require authentication.
func (c *Client) ListCompanies(ctx context.Context) ([]*Company, error) {
	url := c.baseURL + "/open/companies.json"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req = req.WithContext(ctx)

	// Execute request (no authentication required for companies endpoint)
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute http request: %w", err)
	}
	defer res.Body.Close()

	// Read response body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	companies := []*Company{}
	err = json.Unmarshal(body, &companies)
	if err != nil {
		return nil, fmt.Errorf("failed to parse json response: %w", err)
	}

	return companies, nil
}

// doWithToken executes an HTTP request with authentication token.
// If the request returns a 403 status, it invalidates the cached token,
// fetches a new one, and retries the request once.
func (c *Client) doWithToken(req *http.Request) (*http.Response, error) {
	// Attempt 1: Get token (cached or fresh) and execute request
	token, err := c.getToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}
	req.Header.Set("x-api-token", token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute http request: %w", err)
	}

	// If not 403, return the response as-is (caller handles status codes)
	if res.StatusCode != http.StatusForbidden {
		return res, nil
	}

	// 403 detected - close current response body and retry with fresh token
	res.Body.Close()

	// Invalidate cached token and fetch fresh one
	c.invalidateCachedToken()
	token, err = c.getToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get fresh token after 403: %w", err)
	}
	req.Header.Set("x-api-token", token)

	// Attempt 2: Retry request with fresh token
	res, err = c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute http request on retry: %w", err)
	}

	return res, nil
}

// getToken retrieves a cached token or fetches a new one from TokenClient.
// Uses double-checked locking to ensure only one goroutine fetches a new token.
func (c *Client) getToken() (string, error) {
	// Fast path: check for cached token with read lock
	c.mu.RLock()
	token := c.token
	c.mu.RUnlock()

	if token != "" {
		return token, nil
	}

	// Slow path: acquire write lock to fetch new token
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check: another goroutine might have fetched it while we waited for the lock
	if c.token != "" {
		return c.token, nil
	}

	// Fetch new token while holding write lock (blocks other goroutines)
	token, err := c.tokenProvider.GetToken(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}

	// Cache the new token
	c.token = token

	return token, nil
}

// invalidateCachedToken clears the cached token, forcing a fresh token fetch on next request
func (c *Client) invalidateCachedToken() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = ""
}
