package media

import (
	"net/http"
	"slices"
	"strings"
)

var ALLOWED_PROXY_HEADERS = []string{
	"accept",
	"accept-language",
	"accept-encoding",
	"accept-ranges",
	"range",
	"content-length",
	"content-range",
	"content-type",
	"last-modified",
}

func FilterProxyHeaders(headers http.Header) http.Header {
	result := http.Header{}

	for key, values := range headers {
		lowerKey := strings.ToLower(key)
		if !slices.Contains(ALLOWED_PROXY_HEADERS, lowerKey) {
			continue
		}

		for _, value := range values {
			result.Add(key, value)
		}
	}

	return result
}
