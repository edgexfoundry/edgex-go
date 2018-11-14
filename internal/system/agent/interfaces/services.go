package interfaces

// TODO: The abstraction which should be accessed via a global var.

// ServiceStopper is able to stop services
type ServiceStopper interface {
	// Stop stops the service
	Stop(service string, params []string) error
}

// ServiceStarter is able to start services
type ServiceStarter interface {
	// Start starts the service
	Start(service string, params []string) error
}

// ServiceRestarter is able to restart services
type ServiceRestarter interface {
	// Restart starts the service
	Restart(service string, params []string) error
}

// ServiceEnabler is able to enable services to auto-start on boot
type ServiceEnabler interface {
	// Enable enables the service to be automatically started on boot
	Enable(service string, params []string) error
}

// ServiceDisabler is able to disable services, preventing them from auto-starting on boot
type ServiceDisabler interface {
	// Disable disables the service from automatically starting on boot
	Disable(service string, params []string) error
}
