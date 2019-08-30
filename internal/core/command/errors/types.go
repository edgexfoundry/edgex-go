package errors

import "fmt"

type ErrDeviceLocked struct {
	device string
}

func (e ErrDeviceLocked) Error() string {
	return fmt.Sprintf("device '%s' is in admin locked state", e.device)
}

func NewErrDeviceLocked(name string) error {
	return ErrDeviceLocked{device: name}
}
