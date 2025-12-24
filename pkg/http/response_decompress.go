package http

import (
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"

	"github.com/andybalholm/brotli"
)

// DecompressResponse automatically check the compression of response and return's as io.Reader
func DecompressResponse(r *http.Response) (io.ReadCloser, error) {
	encoding := r.Header.Get("Content-Encoding")

	switch encoding {
	case "deflate":
		return flate.NewReader(r.Body), nil

	case "br":
		return io.NopCloser(brotli.NewReader(r.Body)), nil

	case "gzip":
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			return nil, fmt.Errorf("error creating gzip reader: %w", err)
		}
		return gz, nil

	case "", "identity":
		return r.Body, nil

	default:
		return nil, fmt.Errorf("unsupported content encoding: %s", encoding)
	}
}
