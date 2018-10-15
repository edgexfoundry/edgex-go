package agent

type MSMClient interface {
	ProcessResponse(response string) ConfigRespMap
}
