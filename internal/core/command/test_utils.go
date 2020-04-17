package command

import (
	"net/http"
	"strings"
)

// newRequestWithHeaders accepts httpMethod and outlines, which is a map[string]string,
// and, by looping over the entries in outlines, populates the HTTP headers of the
// http.Request, which is returned from this function.
func newRequestWithHeaders(outlines map[string]string, httpMethod string) *http.Request {
	req, _ := http.NewRequest(httpMethod, "/", strings.NewReader("Test body"))
	for k, v := range outlines {
		req.Header.Set(k, v)
	}

	return req
}
