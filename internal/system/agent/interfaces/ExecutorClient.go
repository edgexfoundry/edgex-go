package interfaces

// TODO: The abstraction which should be accessed via a global var.

type ExecutorClient interface {
	StopService(service string) error
}
