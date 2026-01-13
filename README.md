# CSG Medicare Supplement API Client

Go client library for the CSG Actuarial Medicare Supplement API.

## Installation

```bash
go get github.com/medicareschoolcom/csg
```

## Overview

This package provides a Go client for interacting with the CSG Actuarial Medicare Supplement quotes API. It includes:

- Type-safe client for Medicare Supplement quote retrieval
- Support for all standard Medicare Supplement plans (F, G, N)
- Automatic token management with caching and retry logic
- Companies listing endpoint

## Usage

### Creating a Client

Create a client by providing an HTTP client and a token provider:

```go
package main

import (
    "context"
    "net/http"
    "time"

    "github.com/medicareschoolcom/csg/medsup"
)

func main() {
    httpClient := &http.Client{
        Timeout: 30 * time.Second,
    }

    // Implement the TokenProvider interface
    tokenProvider := &YourTokenProvider{}

    client := medsup.NewClient(httpClient, tokenProvider)
}
```

### Implementing TokenProvider

The client requires a `TokenProvider` to handle authentication:

```go
type TokenProvider interface {
    GetToken(ctx context.Context) (string, error)
}
```

Implement this interface to provide your authentication token:

```go
type MyTokenProvider struct {
    // your fields
}

func (tp *MyTokenProvider) GetToken(ctx context.Context) (string, error) {
    // your token retrieval logic
    return "your-api-token", nil
}
```

### Custom Base URL

You can override the default base URL:

```go
client := medsup.NewClient(httpClient, tokenProvider,
    medsup.WithBaseURL("https://custom-api.example.com/v1/med_supp"))
```

### Getting Quotes

```go
effectiveDate := time.Now().Add(30 * 24 * time.Hour)
params := medsup.ListQuotesParams{
    Zip5:          "90210",
    Counties:      []string{"LOS ANGELES"},
    Age:           65,
    Gender:        medsup.Male,
    Tobacco:       medsup.False,
    Plan:          medsup.PlanG,
    EffectiveDate: &effectiveDate,
}

quotes, err := client.ListQuotes(ctx, params)
if err != nil {
    log.Fatal(err)
}

for _, quote := range quotes {
    fmt.Printf("%s: $%d/month\n",
        quote.CompanyBase.Name,
        quote.Rate.Month)
}
```

### Listing Available Companies

```go
companies, err := client.ListCompanies(ctx)
if err != nil {
    log.Fatal(err)
}

for _, company := range companies {
    fmt.Printf("%s (NAIC: %s)\n", company.Name, company.NAIC)
}
```

Note: The `ListCompanies` endpoint does not require authentication.

## API Reference

### Client

```go
func NewClient(httpClient *http.Client, tokenProvider TokenProvider, options ...Option) *Client
```

Creates a new Medicare Supplement quotes client.

**Options:**
- `WithBaseURL(baseURL string)` - Configure a custom base URL

### Methods

#### ListQuotes

```go
func (c *Client) ListQuotes(ctx context.Context, params ListQuotesParams) ([]*Quote, error)
```

Retrieves Medicare Supplement quotes from CSG API.

**Parameters:**
- `Zip5` (required) - 5-digit ZIP code
- `Counties` (optional) - List of county names
- `Age` (required) - Age of the insured (e.g., 65)
- `Gender` (required) - `medsup.Male` or `medsup.Female`
- `Tobacco` (required) - `medsup.True` or `medsup.False`
- `Select` (optional) - Select plan indicator
- `NAICS` (optional) - List of NAIC codes to filter companies
- `Plan` (required) - Plan type: `medsup.PlanF`, `medsup.PlanG`, or `medsup.PlanN`
- `EffectiveDate` (optional) - Policy effective date (defaults to today)

#### ListCompanies

```go
func (c *Client) ListCompanies(ctx context.Context) ([]*Company, error)
```

Retrieves the list of available insurance companies. This endpoint does not require authentication.

### Types

#### Quote
```go
type Quote struct {
    Age         int       // Age of insured
    Gender      string    // Gender: "M" or "F"
    Plan        string    // Plan type: "F", "G", or "N"
    Tobacco     bool      // Tobacco user
    Rate        QuoteRate // Premium rates
    CompanyBase Company   // Insurance company info
}
```

#### QuoteRate
```go
type QuoteRate struct {
    Quarter    int  // Quarterly premium in cents
    Annual     int  // Annual premium in cents
    SemiAnnual int  // Semi-annual premium in cents
    Month      int  // Monthly premium in cents
}
```

#### Company
```go
type Company struct {
    NameFull string  // Full company name
    Name     string  // Short company name
    NAIC     string  // NAIC code
}
```

### Constants

#### Gender
```go
const (
    Male   Gender = "M"
    Female Gender = "F"
)
```

#### Boolean
```go
const (
    True  Boolean = "1"
    False Boolean = "0"
)
```

#### Plan
```go
const (
    PlanF Plan = "F"
    PlanG Plan = "G"
    PlanN Plan = "N"
)
```

## Features

- **Automatic Token Management**: The client automatically caches authentication tokens and refreshes them when they expire (on 403 responses)
- **Thread-Safe**: Token caching uses double-checked locking for safe concurrent access
- **Context Support**: All API methods accept context for cancellation and timeout control
- **Type Safety**: Strong typing for all parameters and responses

## Testing

Run unit and integration tests:

```bash
# Run all tests
go test ./...

# Skip integration tests
go test -short ./...

# Run integration tests
export TEST_API_TOKEN="your-api-token"
go test ./medsup -v
```

Integration tests require:
- Network connectivity
- Valid CSG API credentials (set via `TEST_API_TOKEN` environment variable)
- Access to `api.csgactuarial.com`

## Requirements

- Go 1.24.3 or later
- Valid CSG API credentials (for authenticated endpoints)
- Network connectivity to `api.csgactuarial.com`
