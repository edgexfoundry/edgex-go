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

// ErrExtractingInfoFromRequest is a struct that serves as the value
// receiver for Error as defined for NewErrExtractingInfoFromRequest
type ErrExtractingInfoFromRequest struct {
}

// Error returns a meaningful string message describing error details.
func (e ErrExtractingInfoFromRequest) Error() string {
	return fmt.Sprintf("error extracting command id and device id.")
}

// NewErrExtractingInfoFromRequest returns the relevant, properly-
// constructed error type.
func NewErrExtractingInfoFromRequest() error {
	return ErrExtractingInfoFromRequest{}
}

// ErrBadRequest is a struct that serves as the value receiver
// for Error as defined for NewErrParsingOriginalRequest
type ErrBadRequest struct {
	value string
}

// Error returns a meaningful string message describing error details.
func (e ErrBadRequest) Error() string {
	return fmt.Sprintf("error in parsing related to '%s'", e.value)
}

// NewErrParsingOriginalRequest returns the relevant, properly-
// constructed error type.
func NewErrParsingOriginalRequest(invalid string) error {
	return ErrBadRequest{value: invalid}
}
