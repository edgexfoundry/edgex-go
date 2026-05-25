package inspect

type SdkInspectResponse struct {
	Errors  []string       `json:"errors"`
	Success bool           `json:"success"`
	Values  map[string]any `json:"values"`
}

type RouterConnInspectDetail struct {
	RouterName   string               `json:"routerName"`
	RouterAddr   string               `json:"routerAddr"`
	Closed       bool                 `json:"closed"`
	VirtualConns []*VirtualConnDetail `json:"virtualConns"`
}

type VirtualConnDetail struct {
	ConnId      uint32 `json:"connId"`
	SinkType    string `json:"sinkType"`
	ServiceName string `json:"serviceName"`
	Closed      bool   `json:"closed"`
	CircuitId   string `json:"circuitId,omitempty"`
}

type ContextInspectResult struct {
	ContextId         string                     `json:"contextId"`
	Identity          *ContextInspectIdentity    `json:"identity"`
	Services          []*ContextInspectService   `json:"services"`
	Sessions          []*ContextInspectSession   `json:"sessions"`
	RouterConnections []*RouterConnInspectDetail `json:"routerConnections"`
	Listeners         []*ContextInspectListener  `json:"listeners"`
}

type ContextInspectIdentity struct {
	Id   string `json:"id"`
	Name string `json:"name,omitempty"`
}

type ContextInspectService struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

type ContextInspectSession struct {
	Id        string `json:"id"`
	ServiceId string `json:"serviceId"`
	Type      string `json:"type"`
}

type ContextInspectListener struct {
	ServiceId      string `json:"serviceId"`
	ServiceName    string `json:"serviceName"`
	MaxTerminators int    `json:"maxTerminators"`
	ListenerCount  int    `json:"listenerCount"`
}
