// Copyright 2021 Contributors to the Parsec project.
// SPDX-License-Identifier: Apache-2.0

package requests

import (
	"bytes"
	"reflect"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

const wireHeaderSizeValue uint16 = 30
const WireHeaderSize uint16 = wireHeaderSizeValue + 6

// StatusCode type to represent status codes in response headers
type StatusCode uint16

// StatusCode values for response status codes defined here: https://parallaxsecond.github.io/parsec-book/parsec_client/status_codes.html.
const (
	// Service Internal Response Status Codes
	StatusSuccess                         StatusCode = 0  // Operation was a success
	StatusWrongProviderID                 StatusCode = 1  // Requested provider ID does not match that of the backend
	StatusContentTypeNotSupported         StatusCode = 2  // Requested content type is not supported by the backend
	StatusAcceptTypeNotSupported          StatusCode = 3  // Requested accept type is not supported by the backend
	StatusWireProtocolVersionNotSupported StatusCode = 4  // Requested version is not supported by the backend
	StatusProviderNotRegistered           StatusCode = 5  // No provider registered for the requested provider ID
	StatusProviderDoesNotExist            StatusCode = 6  // No provider defined for requested provider ID
	StatusDeserializingBodyFailed         StatusCode = 7  // Failed to deserialize the body of the message
	StatusSerializingBodyFailed           StatusCode = 8  // Failed to serialize the body of the message
	StatusOpcodeDoesNotExist              StatusCode = 9  // Requested operation is not defined
	StatusResponseTooLarge                StatusCode = 10 // Response size exceeds allowed limits
	StatusAuthenticationError             StatusCode = 11 // Authentication failed
	StatusAuthenticatorDoesNotExist       StatusCode = 12 // Authenticator not supported
	StatusAuthenticatorNotRegistered      StatusCode = 13 // Authenticator not supported
	StatusKeyInfoManagerError             StatusCode = 14 // Internal error in the Key Info Manager
	StatusConnectionError                 StatusCode = 15 // Generic input/output error
	StatusInvalidEncoding                 StatusCode = 16 // Invalid value for this data type
	StatusInvalidHeader                   StatusCode = 17 // Constant fields in header are invalid
	StatusWrongProviderUUID               StatusCode = 18 // The UUID vector needs to only contain 16 bytes
	StatusNotAuthenticated                StatusCode = 19 // Request did not provide a required authentication
	StatusBodySizeExceedsLimit            StatusCode = 20 // Request length specified in the header is above defined limit
	StatusAdminOperation                  StatusCode = 21 // The operation requires admin privilege

	// PSA Response Status Codes

	StatusPsaErrorGenericError         StatusCode = 1132 // An error occurred that does not correspond to any defined failure cause
	StatusPsaErrorNotPermitted         StatusCode = 1133 // The requested action is denied by a policy
	StatusPsaErrorNotSupported         StatusCode = 1134 // The requested operation or a parameter is not supported by this implementation
	StatusPsaErrorInvalidArgument      StatusCode = 1135 // The parameters passed to the function are invalid
	StatusPsaErrorInvalidHandle        StatusCode = 1136 // The key handle is not valid
	StatusPsaErrorBadState             StatusCode = 1137 // The requested action cannot be performed in the current state
	StatusPsaErrorBufferTooSmall       StatusCode = 1138 // An output buffer is too small
	StatusPsaErrorAlreadyExists        StatusCode = 1139 // Asking for an item that already exists
	StatusPsaErrorDoesNotExist         StatusCode = 1140 // Asking for an item that doesn't exist
	StatusPsaErrorInsufficientMemory   StatusCode = 1141 // There is not enough runtime memory
	StatusPsaErrorInsufficientStorage  StatusCode = 1142 // There is not enough persistent storage available
	StatusPsaErrorInssuficientData     StatusCode = 1143 // Insufficient data when attempting to read from a resource
	StatusPsaErrorCommunicationFailure StatusCode = 1145 // There was a communication failure inside the implementation
	StatusPsaErrorStorageFailure       StatusCode = 1146 // There was a storage failure that may have led to data loss
	StatusPsaErrorHardwareFailure      StatusCode = 1147 // A hardware failure was detected
	StatusPsaErrorInsufficientEntropy  StatusCode = 1148 // There is not enough entropy to generate random data needed for the requested action
	StatusPsaErrorInvalidSignature     StatusCode = 1149 // The signature, MAC or hash is incorrect
	StatusPsaErrorInvalidPadding       StatusCode = 1150 // The decrypted padding is incorrect
	StatusPsaErrorCorruptionDetected   StatusCode = 1151 // A tampering attempt was detected
	StatusPsaErrorDataCorrupt          StatusCode = 1152 // Stored data has been corrupted

)

func (code StatusCode) IsValid() bool {
	return (code >= StatusSuccess && code <= StatusAdminOperation) || (code >= StatusPsaErrorGenericError && code <= StatusPsaErrorDataCorrupt)
}

// ParseResponse returns a response if it successfully unmarshals the given byte buffer
func ParseResponse(expectedOpCode OpCode, buf *bytes.Buffer, responseProtoBuf proto.Message) error {
	if buf == nil {
		return errors.Errorf("nil buffer supplied")
	}
	if responseProtoBuf == nil || reflect.ValueOf(responseProtoBuf).IsNil() {
		return errors.Errorf("nil message supplied")
	}

	hdrBuf := make([]byte, WireHeaderSize)
	_, err := buf.Read(hdrBuf)
	if err != nil {
		return errors.Wrap(err, "failed to read header")
	}
	wireHeader, err := parseWireHeaderFromBuf(bytes.NewBuffer(hdrBuf))
	if err != nil {
		return errors.Wrap(err, "failed to parse header")
	}
	if wireHeader.opCode != expectedOpCode {
		// If we've not got the opcode we expect, don't even try to deserialise the body.
		return errors.Errorf("was expecting response with op code %v, got %v", expectedOpCode, wireHeader.opCode)
	}

	bodyBuf := make([]byte, wireHeader.bodyLen)
	n, err := buf.Read(bodyBuf)
	if err != nil {
		return errors.Wrap(err, "failed to read body")
	}
	if uint32(n) != wireHeader.bodyLen {
		return errors.Errorf("body underflow error, expected %v bytes, got %v", wireHeader.bodyLen, n)
	}
	err = proto.Unmarshal(bodyBuf, responseProtoBuf)
	if err != nil {
		return err
	}

	return wireHeader.Status.ToErr()
}

// ToErr returns nil if the response code is a success, or an appropriate error otherwise.
//
//nolint:gocyclo
func (code StatusCode) ToErr() error {
	switch code {
	case StatusSuccess:
		return nil
	case StatusWrongProviderID:
		return errors.Errorf("wrong provider id")
	case StatusContentTypeNotSupported:
		return errors.Errorf("content type not supported")
	case StatusAcceptTypeNotSupported:
		return errors.Errorf("accept type not supported")
	case StatusWireProtocolVersionNotSupported:
		return errors.Errorf("requested version is not supported by the backend")
	case StatusProviderNotRegistered:
		return errors.Errorf("provider not registered")
	case StatusProviderDoesNotExist:
		return errors.Errorf("provider does not exist")
	case StatusDeserializingBodyFailed:
		return errors.Errorf("deserializing body failed")
	case StatusSerializingBodyFailed:
		return errors.Errorf("serializing body failed")
	case StatusOpcodeDoesNotExist:
		return errors.Errorf("opcode does not exist")
	case StatusResponseTooLarge:
		return errors.Errorf("response too large")
	case StatusAuthenticationError:
		return errors.Errorf("authentication error")
	case StatusAuthenticatorDoesNotExist:
		return errors.Errorf("authentication does not exist")
	case StatusAuthenticatorNotRegistered:
		return errors.Errorf("authentication not registered")
	case StatusKeyInfoManagerError:
		return errors.Errorf("internal error in the Key Info Manager")
	case StatusConnectionError:
		return errors.Errorf("generic input/output error")
	case StatusInvalidEncoding:
		return errors.Errorf("invalid value for this data type")
	case StatusInvalidHeader:
		return errors.Errorf("constant fields in header are invalid")
	case StatusWrongProviderUUID:
		return errors.Errorf("the UUID vector needs to only contain 16 bytes")
	case StatusNotAuthenticated:
		return errors.Errorf("request did not provide a required authentication")
	case StatusBodySizeExceedsLimit:
		return errors.Errorf("request length specified in the header is above defined limit")
	case StatusAdminOperation:
		return errors.Errorf("the operation requires admin privilege")

	case StatusPsaErrorGenericError:
		return errors.Errorf("generic error")
	case StatusPsaErrorNotPermitted:
		return errors.Errorf("not permitted")
	case StatusPsaErrorNotSupported:
		return errors.Errorf("not supported")
	case StatusPsaErrorInvalidArgument:
		return errors.Errorf("invalid argument")
	case StatusPsaErrorInvalidHandle:
		return errors.Errorf("invalid handle")
	case StatusPsaErrorBadState:
		return errors.Errorf("bad state")
	case StatusPsaErrorBufferTooSmall:
		return errors.Errorf("buffer too small")
	case StatusPsaErrorAlreadyExists:
		return errors.Errorf("already exists")
	case StatusPsaErrorDoesNotExist:
		return errors.Errorf("does not exist")
	case StatusPsaErrorInsufficientMemory:
		return errors.Errorf("insufficient memory")
	case StatusPsaErrorInsufficientStorage:
		return errors.Errorf("insufficient storage")
	case StatusPsaErrorInssuficientData:
		return errors.Errorf("insufficient data")
	case StatusPsaErrorCommunicationFailure:
		return errors.Errorf("communications failure")
	case StatusPsaErrorStorageFailure:
		return errors.Errorf("storage failure")
	case StatusPsaErrorHardwareFailure:
		return errors.Errorf("hardware failure")
	case StatusPsaErrorInsufficientEntropy:
		return errors.Errorf("insufficient entropy")
	case StatusPsaErrorInvalidSignature:
		return errors.Errorf("invalid signature")
	case StatusPsaErrorInvalidPadding:
		return errors.Errorf("invalid padding")
	case StatusPsaErrorCorruptionDetected:
		return errors.Errorf("tampering detected")
	case StatusPsaErrorDataCorrupt:
		return errors.Errorf("stored data has been corrupted")
	}
	return errors.Errorf("unknown error code")
}
