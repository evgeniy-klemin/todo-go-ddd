package ports

import (
	"fmt"
	"net/http"
	"strings"
)

func cursorLinks(req *http.Request, perPage int, hasNext bool, nextCursorEncoded string) string {
	var parts []string

	if hasNext {
		q := req.URL.Query()
		q.Set("_cursor", nextCursorEncoded)
		parts = append(parts, fmt.Sprintf("<%s?%s>;rel=next", req.URL.Path, q.Encode()))
	}

	// rel=first — no cursor param
	q := req.URL.Query()
	q.Del("_cursor")
	firstURL := req.URL.Path
	if encoded := q.Encode(); encoded != "" {
		firstURL += "?" + encoded
	}
	parts = append(parts, fmt.Sprintf("<%s>;rel=first", firstURL))

	return strings.Join(parts, ",")
}
