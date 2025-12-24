package oauth2

import (
	"context"
	"net/http"
)

// OAuth2TokenManagerProvider defines the contract for managing OAuth 2.0 token acquisition
// and authentication. It provides methods to configure token requests and apply authentication
// to HTTP requests.
type OAuth2TokenManagerProvider interface {
	// SendAsPost configures the token manager to send token requests using the HTTP POST method.
	// This is the most common method for OAuth 2.0 token requests as recommended by RFC 6749.
	SendAsPost()

	// SendAsGet configures the token manager to send token requests using the HTTP GET method.
	// Note: While supported by some providers, POST is generally preferred for security reasons
	// as credentials are sent in the request body rather than the URL.
	SendAsGet()

	// WithOptionalParams adds optional parameters to the token request.
	// These parameters will be included in the token request according to the configured
	// HTTP method (POST body or GET query parameters).
	//
	// Common optional parameters include:
	//   - resource: The target resource for the token
	//   - audience: The intended audience of the token
	//   - custom provider-specific parameters
	//
	// Parameters:
	//   - params: A map of parameter names to values
	WithOptionalParams(params map[string]string)

	// WithAuthenticationURL sets the OAuth 2.0 token endpoint URL.
	// This is the URL where token requests will be sent (e.g., "https://provider.com/oauth/token").
	//
	// Parameters:
	//   - url: The complete URL of the OAuth 2.0 token endpoint
	WithAuthenticationURL(url string)

	// SetAuthorizationHeader obtains an access token for the specified scope and sets it
	// in the Authorization header of the provided HTTP request.
	//
	// This method handles the complete token lifecycle:
	//   1. Checks for a valid cached token
	//   2. If no valid token exists, requests a new token from the authentication URL
	//   3. Sets the "Authorization: Bearer <token>" header on the request
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//   - r: The HTTP request to authenticate
	//   - scope: The OAuth 2.0 scope(s) to request (space-separated if multiple)
	//
	// Returns:
	//   - error: An error if token acquisition or header setting fails
	SetAuthorizationHeader(ctx context.Context, r *http.Request, scope string) error
}
