package internal

import "net/http"

type HttpCaller interface {
	Do(req *http.Request) (*http.Response, error)
}
