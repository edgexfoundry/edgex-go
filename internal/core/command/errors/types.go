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

type ErrCommandNotAssociatedWithDevice struct {
	commandID string
	deviceID  string
}

func (e ErrCommandNotAssociatedWithDevice) Error() string {
	return fmt.Sprintf("Command with id '%v' does not belong to device with id '%v'.", e.commandID, e.deviceID)
}

func NewErrCommandNotAssociatedWithDevice(commandID string, deviceID string) error {
	return ErrCommandNotAssociatedWithDevice{commandID, deviceID}
}
