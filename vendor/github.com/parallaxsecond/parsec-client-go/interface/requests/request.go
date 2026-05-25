// Copyright 2021 Contributors to the Parsec project.
// SPDX-License-Identifier: Apache-2.0

package requests

import (
	"bytes"

	"github.com/parallaxsecond/parsec-client-go/interface/auth"
	"google.golang.org/protobuf/proto"
)

const requestHeaderSize uint16 = 30

// RequestBody represents a marshaled request body
type RequestBody struct {
	*bytes.Buffer
}

// Request represents a Parsec request
type Request struct {
	Header wireHeader
	Body   RequestBody
	Auth   auth.RequestAuthToken
}

// NewRequest creates a new request based on the opcode and the message.
func NewRequest(op OpCode, bdy proto.Message, authenticator auth.Authenticator, provider ProviderID) (*Request, error) {
	bodyBuf, err := proto.Marshal(bdy)
	if err != nil {
		return nil, err
	}

	authtok, err := authenticator.NewRequestAuth()
	if err != nil {
		return nil, err
	}
	r := &Request{
		Header: wireHeader{
			versionMajor: versionMajorOne,
			versionMinor: versionMinorZero,
			flags:        flagsZero,
			provider:     provider,
			// todo set session handles
			contentType: contentTypeProtobuf,
			authType:    authtok.AuthType(),
			bodyLen:     uint32(len(bodyBuf)),
			authLen:     uint16(authtok.Buffer().Len()),
			opCode:      op,
			Status:      StatusSuccess,
		},
		Body: RequestBody{
			bytes.NewBuffer(bodyBuf),
		},
		Auth: authtok,
	}
	return r, nil
}

// Pack encodes a request to the wire format
func (r *Request) Pack() (*bytes.Buffer, error) {
	if err := r.Header.checkForRequest(); err != nil {
		return nil, err
	}
	b := bytes.NewBuffer([]byte{})
	err := r.Header.pack(b)
	if err != nil {
		return nil, err
	}
	_, err = b.Write(r.Body.Bytes())
	if err != nil {
		return nil, err
	}
	_, err = b.Write(r.Auth.Buffer().Bytes())
	if err != nil {
		return nil, err
	}
	return b, nil
}
