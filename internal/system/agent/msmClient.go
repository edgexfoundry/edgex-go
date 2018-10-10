package agent

type MSMClient interface {
	ProcessConfigResponse(response string) ConfigRespMap
	ProcessMetricsResponse(response string) MetricsRespMap
}
