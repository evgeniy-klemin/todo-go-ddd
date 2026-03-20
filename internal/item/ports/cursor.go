package ports

import "encoding/base64"

// encodeCursor base64url-encodes an opaque cursor []byte so it is safe to embed in
// HTTP response headers and JSON bodies. Uses RawURLEncoding (no padding) to produce
// URL-safe output without trailing '=' characters.
func encodeCursor(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// decodeCursor decodes a base64url-encoded cursor string received from an HTTP request
// (e.g. the ?cursor= query parameter) back to the opaque []byte expected by the app layer.
// Returns an error if s is not valid base64url; the caller should respond with HTTP 400.
func decodeCursor(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}
