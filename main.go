package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// HTTPClient is an interface for sending HTTP requests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// HTTPClientFunc is a function type that implements the HTTPClient interface.
type HTTPClientFunc func(req *http.Request) (*http.Response, error)

// Do sends an HTTP request and returns an HTTP response.
func (fn HTTPClientFunc) Do(req *http.Request) (*http.Response, error) {
	return fn(req)
}

// Middleware is a type for functions that modify HTTPClient behavior.
type Middleware func(HTTPClient) HTTPClient

// BasicAuthMiddleware adds Basic Auth to the request.
func BasicAuthMiddleware(username, password string) Middleware {
	return func(client HTTPClient) HTTPClient {
		return HTTPClientFunc(func(req *http.Request) (*http.Response, error) {
			req.SetBasicAuth(username, password)
			return client.Do(req)
		})
	}
}

// APIKeyAuthMiddleware adds API key-based authentication to the request.
func APIKeyAuthMiddleware(apiKey string) Middleware {
	return func(client HTTPClient) HTTPClient {
		return HTTPClientFunc(func(req *http.Request) (*http.Response, error) {
			req.Header.Set("Authorization", "Bearer "+apiKey)
			return client.Do(req)
		})
	}
}

// CustomClient is a custom HTTP client with middleware support.
type CustomClient struct {
	httpClient HTTPClient
}

// NewCustomClient creates a new CustomClient with optional middleware.
func NewCustomClient(baseClient HTTPClient, middlewares ...Middleware) CustomClient {
	// Apply middleware to the base HTTP client.
	for _, middleware := range middlewares {
		baseClient = middleware(baseClient)
	}
	return CustomClient{httpClient: baseClient}
}

// Get sends a GET request and returns the response body.
func (c *CustomClient) Get(ctx context.Context, url string) ([]byte, error) {
	// Create a new GET request with context.
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Send the request using the custom HTTP client.
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read and return the response body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

func main() {
	// Define your API key and endpoint.
	apiKey := "your-api-key-here"
	apiEndpoint := "https://your-api-endpoint.com"

	// Create a new custom client with API key authentication middleware.
	client := NewCustomClient(http.DefaultClient, APIKeyAuthMiddleware(apiKey))

	// Send a GET request and print the response body.
	responseBody, err := client.Get(context.Background(), apiEndpoint)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(string(responseBody))
}
