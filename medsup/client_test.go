package medsup

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"
)

// TestClient_ListQuotes_Integration tests the ListQuotes method with real API calls.
// This integration test verifies end-to-end functionality including:
// - Token acquisition from inhouse API
// - CSG API quote request
// - Response parsing and validation
//
// Requirements:
// - Network connectivity
// - Access to medicare-school-quote-tool.herokuapp.com
// - Valid CSG API credentials
//
// Skip with: go test -short
func TestClient_ListQuotes_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	token := os.Getenv("TEST_API_TOKEN")
	if token == "" {
		t.Skip("set TEST_API_TOKEN to run this test")
	}

	// Setup: Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Setup: Create token provider
	tokenProvider := &testTokenProvider{token: token}

	// Setup: Create client
	client := NewClient(httpClient, tokenProvider)

	// Setup: Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Setup: Prepare test parameters
	effectiveDate := time.Now().Add(30 * 24 * time.Hour) // 30 days from now
	params := ListQuotesParams{
		Zip5:          "90210",
		Counties:      []string{"LOS ANGELES"},
		Age:           65,
		Gender:        Male,
		Tobacco:       False,
		Plan:          PlanG,
		EffectiveDate: &effectiveDate,
	}

	// Execute: Call ListQuotes
	quotes, err := client.ListQuotes(ctx, params)
	if err != nil {
		t.Fatalf("ListQuotes returned error: %v", err)
	}

	// Assert: Quotes should not be nil
	if quotes == nil {
		t.Fatalf("Expected non-nil quotes slice, got nil")
	}

	// Assert: Should have at least one quote
	if len(quotes) == 0 {
		t.Fatalf("Expected at least one quote, got empty slice")
	}

	// Assert: Validate first quote structure
	firstQuote := quotes[0]

	if firstQuote.Age != 65 {
		t.Errorf("Expected age 65, got %d", firstQuote.Age)
	}

	if firstQuote.Gender != "M" {
		t.Errorf("Expected gender 'M', got %s", firstQuote.Gender)
	}

	if firstQuote.Plan != "G" {
		t.Errorf("Expected plan 'G', got %s", firstQuote.Plan)
	}

	// Assert: Rate should have valid positive values
	if firstQuote.Rate.Month <= 0 {
		t.Errorf("Expected positive monthly rate, got %d", firstQuote.Rate.Month)
	}

	if firstQuote.Rate.Quarter <= 0 {
		t.Errorf("Expected positive quarterly rate, got %d", firstQuote.Rate.Quarter)
	}

	if firstQuote.Rate.SemiAnnual <= 0 {
		t.Errorf("Expected positive semi-annual rate, got %d", firstQuote.Rate.SemiAnnual)
	}

	if firstQuote.Rate.Annual <= 0 {
		t.Errorf("Expected positive annual rate, got %d", firstQuote.Rate.Annual)
	}

	// Assert: Company base should have non-empty values
	if firstQuote.CompanyBase.Name == "" {
		t.Errorf("Expected non-empty company name, got empty string")
	}

	if firstQuote.CompanyBase.NameFull == "" {
		t.Errorf("Expected non-empty company full name, got empty string")
	}

	if firstQuote.CompanyBase.NAIC == "" {
		t.Errorf("Expected non-empty NAIC code, got empty string")
	}

	// Success: Log summary
	t.Logf("Successfully retrieved %d quotes", len(quotes))
	t.Logf("First quote: %s - $%d/month", firstQuote.CompanyBase.Name, firstQuote.Rate.Month)
}

// TestClient_ListCompanies_Integration tests the ListCompanies method with real API calls.
// This integration test verifies end-to-end functionality including:
// - CSG API companies request (unauthenticated endpoint)
// - Response parsing and validation
//
// Requirements:
// - Network connectivity
// - Access to api.csgactuarial.com
//
// Skip with: go test -short
func TestClient_ListCompanies_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup: Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Setup: Create client (no token provider needed for this endpoint)
	client := NewClient(httpClient, nil)

	// Setup: Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Execute: Call ListCompanies
	companies, err := client.ListCompanies(ctx)
	if err != nil {
		t.Fatalf("ListCompanies returned error: %v", err)
	}

	// Assert: Companies should not be nil
	if companies == nil {
		t.Fatalf("Expected non-nil companies slice, got nil")
	}

	// Assert: Should have at least one company
	if len(companies) == 0 {
		t.Fatalf("Expected at least one company, got empty slice")
	}

	// Assert: Validate first company structure
	firstCompany := companies[0]

	if firstCompany.Name == "" {
		t.Errorf("Expected non-empty company name, got empty string")
	}

	if firstCompany.NameFull == "" {
		t.Errorf("Expected non-empty company full name, got empty string")
	}

	if firstCompany.NAIC == "" {
		t.Errorf("Expected non-empty NAIC code, got empty string")
	}

	// Success: Log summary
	t.Logf("Successfully retrieved %d companies", len(companies))
	t.Logf("First company: %s (NAIC: %s)", firstCompany.Name, firstCompany.NAIC)
	t.Logf("Full name: %s", firstCompany.NameFull)
}

type testTokenProvider struct {
	token string
}

func (tp *testTokenProvider) GetToken(ctx context.Context) (string, error) {
	return tp.token, nil
}
