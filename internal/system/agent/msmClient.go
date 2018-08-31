package agent

import "net/http"

type MSMClient interface {
	ProcessResponse(resp *http.Response) RespMap
}
