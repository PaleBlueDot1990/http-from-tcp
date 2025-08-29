package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
)

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := make(headers.Headers)
	h["Content-Length"] = fmt.Sprintf("%d", contentLen)
	h["Connection"] = "close"
	h["Content-Type"] = "text/plain"
	return h
}
