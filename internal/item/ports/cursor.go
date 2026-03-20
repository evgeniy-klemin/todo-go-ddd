package ports

import "encoding/base64"

// encodeCursor base64-encodes an opaque cursor []byte for use in HTTP responses.
func encodeCursor(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// decodeCursor base64-decodes a cursor string received from an HTTP request.
func decodeCursor(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}
