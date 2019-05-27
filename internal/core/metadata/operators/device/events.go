package device

type DeviceEvent struct {
	DeviceId   string
	DeviceName string
	Error      error
	HttpMethod string
	ServiceId  string
}
