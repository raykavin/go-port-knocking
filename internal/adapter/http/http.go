package http

import (
	"context"
	"mime/multipart"
	"net/http"
)

// RequestContext encapsulates the HTTP request/response cycle and provides
// methods for handling HTTP operations.
type RequestContext interface {
	// Context returns the current context associated with the request.
	Context() context.Context

	// Writer returns the http.ResponseWriter for writing the response.
	Writer() http.ResponseWriter

	// ServeFile serves a file for download
	ServeFile(filename, contentType string, contentBytes []byte) error

	// Request returns the original http.Request.
	Request() *http.Request

	// Set stores a value in the request context with the provided key.
	Set(key string, value any)

	// Get retrieves a value from the request context by key.
	// Returns the value and a boolean indicating if the key exists.
	Get(key string) (any, bool)

	// JSON sends a JSON response with the specified status code.
	JSON(statusCode int, data any)

	// BindQuery parses the query into a provided struct pointer
	BindQuery(dest any) error

	// BindJSON parses the request body as JSON into the provided struct pointer.
	BindJSON(dest any) error

	// GetParam returns the value of the URL parameter with the specified key.
	GetParam(key string) string

	// GetQuery returns the value of the first query parameter with the specified key.
	GetQuery(key string) string

	// GetClientIP returns the value of the client ip address
	GetClientIP() string

	// FormFile returns the file header from a specific key or the default "file" key value
	FormFile(key ...string) (*multipart.FileHeader, error)

	// Redirect sends an HTTP redirect to the specified URL with the given status code.
	Redirect(statusCode int, to string)

	// SetCookie adds an HTTP cookie to the response.
	SetCookie(c *http.Cookie)

	// Abort stops the current request handling chain.
	Abort()

	// StatusCode returns the writer status code
	StatusCode() int

	// Next continues execution to the next handler in the chain.
	Next()

	// GetIDParam parse and returns a id parameter value or error
	GetIDParam() (uint64, error)
}

// Server represents an HTTP server that can be started and shut down.
type Server interface {
	// Listen starts the HTTP server and begins accepting connections.
	// This method blocks until the server is shut down or an error occurs.
	Listen() error

	// Shutdown gracefully shuts down the server without interrupting active connections.
	// It uses the provided context for timeout control.
	Shutdown(ctx context.Context) error

	// GetPort returns a port of HTTP server is running by configuration
	GetPort() uint16

	// IsSSLEnabled returns true if SSL is enabled by configuration
	IsSSLEnabled() bool
}
