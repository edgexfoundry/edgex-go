//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
)

// ErrKind a categorical identifier used to give high-level insight as to the error type.
type ErrKind string

const (
	// Constant Kind identifiers which can be used to label and group errors.
	KindUnknown             ErrKind = "Unknown"
	KindDatabaseError       ErrKind = "Database"
	KindCommunicationError  ErrKind = "Communication"
	KindEntityDoesNotExist  ErrKind = "NotFound"
	KindContractInvalid     ErrKind = "ContractInvalid"
	KindServerError         ErrKind = "UnexpectedServerError"
	KindLimitExceeded       ErrKind = "LimitExceeded"
	KindStatusConflict      ErrKind = "StatusConflict"
	KindDuplicateName       ErrKind = "DuplicateName"
	KindInvalidId           ErrKind = "InvalidId"
	KindServiceUnavailable  ErrKind = "ServiceUnavailable"
	KindNotAllowed          ErrKind = "NotAllowed"
	KindServiceLocked       ErrKind = "ServiceLocked"
	KindNotImplemented      ErrKind = "NotImplemented"
	KindRangeNotSatisfiable ErrKind = "RangeNotSatisfiable"
	KindIOError             ErrKind = "IOError"
	KindOverflowError       ErrKind = "OverflowError"
	KindNaNError            ErrKind = "NaNError"
)

// EdgeX provides an abstraction for all internal EdgeX errors.
// This exists so that we can use this type in our method signatures and return nil which will fit better with the way
// the Go builtin errors are normally handled.
type EdgeX interface {
	// Error obtains the error message associated with the error.
	Error() string
	// DebugMessages returns a detailed string for debug purpose.
	DebugMessages() string
	// Message returns the first level error message without further details.
	Message() string
	// Code returns the status code of this error.
	Code() int
}

// CommonEdgeX generalizes an error structure which can be used for any type of EdgeX error.
type CommonEdgeX struct {
	// callerInfo contains information of function call stacks.
	callerInfo string
	// kind contains information regarding the high level error type.
	kind ErrKind
	// message contains detailed information about the error.
	message string
	// code is the status code to represent this error.
	// We are using the standard HTTP status code: https://www.w3.org/Protocols/rfc2616/rfc2616-sec10.html
	code int
	// err is a nested error which is used to form a chain of errors for better context.
	err error
}

// Kind determines the Kind associated with an error by inspecting the chain of errors. The top-most matching Kind is
// returned or KindUnknown if no Kind can be determined.
func Kind(err error) ErrKind {
	var e CommonEdgeX
	if !errors.As(err, &e) {
		return KindUnknown
	}
	// We want to return the first "Kind" we see that isn't Unknown, because
	// the higher in the stack the Kind was specified the more context we had.
	if e.kind != KindUnknown || e.err == nil {
		return e.kind
	}

	return Kind(e.err)
}

// Error creates an error message taking all nested and wrapped errors into account.
func (ce CommonEdgeX) Error() string {
	if ce.err == nil {
		return ce.message
	}

	// ce.err.Error functionality gets the error message of the nested error and which will handle both CommonEdgeX
	// types and Go standard errors(both wrapped and non-wrapped).
	if ce.message != "" {
		return ce.message + " -> " + ce.err.Error()
	} else {
		return ce.err.Error()
	}
}

// DebugMessages returns a string taking all nested and wrapped operations and errors into account.
func (ce CommonEdgeX) DebugMessages() string {
	if ce.err == nil {
		return ce.callerInfo + ": " + ce.message
	}

	if w, ok := ce.err.(CommonEdgeX); ok {
		return ce.callerInfo + ": " + ce.message + " -> " + w.DebugMessages()
	} else {
		return ce.callerInfo + ": " + ce.message + " -> " + ce.err.Error()
	}
}

// Message returns the first level error message without further details.
func (ce CommonEdgeX) Message() string {
	if ce.message == "" && ce.err != nil {
		if w, ok := ce.err.(CommonEdgeX); ok {
			return w.Message()
		} else {
			return ce.err.Error()
		}
	}

	return ce.message
}

// Code returns the status code of this error.
func (ce CommonEdgeX) Code() int {
	return ce.code
}

// Unwrap retrieves the next nested error in the wrapped error chain.
// This is used by the new wrapping and unwrapping features available in Go 1.13 and aids in traversing the error chain
// of wrapped errors.
func (ce CommonEdgeX) Unwrap() error {
	return ce.err
}

// Is determines if an error is of type CommonEdgeX.
// This is used by the new wrapping and unwrapping features available in Go 1.13 and aids the errors.Is function when
// determining is an error or any error in the wrapped chain contains an error of a particular type.
func (ce CommonEdgeX) Is(err error) bool {
	switch err.(type) {
	case CommonEdgeX:
		return true
	default:
		return false

	}
}

// NewCommonEdgeX creates a new CommonEdgeX with the information provided.
func NewCommonEdgeX(kind ErrKind, message string, wrappedError error) CommonEdgeX {
	return CommonEdgeX{
		kind:       kind,
		callerInfo: getCallerInformation(),
		message:    message,
		code:       codeMapping(kind),
		err:        wrappedError,
	}
}

// NewCommonEdgeXWrapper creates a new CommonEdgeX to wrap another error to record the function call stacks.
func NewCommonEdgeXWrapper(wrappedError error) CommonEdgeX {
	kind := Kind(wrappedError)
	return CommonEdgeX{
		kind:       kind,
		callerInfo: getCallerInformation(),
		message:    "",
		code:       codeMapping(kind),
		err:        wrappedError,
	}
}

// getCallerInformation generates information about the caller function. This function skips the caller which has
// invoked this function, but rather introspects the calling function 3 frames below this frame in the call stack. This
// function is a helper function which eliminates the need for the 'callerInfo' field in the `CommonEdgeX` type and
// providing an 'callerInfo' string when creating an 'CommonEdgeX'
func getCallerInformation() string {
	pc := make([]uintptr, 10)
	runtime.Callers(3, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	return fmt.Sprintf("[%s]-%s(line %d)", file, f.Name(), line)
}

// codeMapping determines the correct HTTP response code for the given error kind.
func codeMapping(kind ErrKind) int {
	switch kind {
	case KindUnknown, KindDatabaseError, KindServerError, KindOverflowError, KindNaNError:
		return http.StatusInternalServerError
	case KindCommunicationError:
		return http.StatusBadGateway
	case KindEntityDoesNotExist:
		return http.StatusNotFound
	case KindContractInvalid, KindInvalidId:
		return http.StatusBadRequest
	case KindStatusConflict, KindDuplicateName:
		return http.StatusConflict
	case KindLimitExceeded:
		return http.StatusRequestEntityTooLarge
	case KindServiceUnavailable:
		return http.StatusServiceUnavailable
	case KindServiceLocked:
		return http.StatusLocked
	case KindNotImplemented:
		return http.StatusNotImplemented
	case KindNotAllowed:
		return http.StatusMethodNotAllowed
	case KindRangeNotSatisfiable:
		return http.StatusRequestedRangeNotSatisfiable
	case KindIOError:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

// KindMapping determines the correct EdgeX error kind for the given HTTP response code.
func KindMapping(code int) ErrKind {
	switch code {
	case http.StatusInternalServerError:
		return KindServerError
	case http.StatusBadGateway:
		return KindCommunicationError
	case http.StatusNotFound:
		return KindEntityDoesNotExist
	case http.StatusBadRequest:
		return KindContractInvalid
	case http.StatusConflict:
		return KindStatusConflict
	case http.StatusRequestEntityTooLarge:
		return KindLimitExceeded
	case http.StatusServiceUnavailable:
		return KindServiceUnavailable
	case http.StatusLocked:
		return KindServiceLocked
	case http.StatusNotImplemented:
		return KindNotImplemented
	case http.StatusMethodNotAllowed:
		return KindNotAllowed
	case http.StatusRequestedRangeNotSatisfiable:
		return KindRangeNotSatisfiable
	default:
		return KindUnknown
	}
}
