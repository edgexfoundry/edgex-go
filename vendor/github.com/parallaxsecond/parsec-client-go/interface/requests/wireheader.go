// Copyright 2021 Contributors to the Parsec project.
// SPDX-License-Identifier: Apache-2.0

package requests

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/parallaxsecond/parsec-client-go/interface/auth"
)

type contentType uint8
type versionMajorType uint8
type versionMinorType uint8
type flagsType uint16
type sessionType uint64
type acceptType uint8

const (
	versionMajorOne  versionMajorType = 0x01
	versionMinorZero versionMinorType = 0x00

	flagsZero flagsType = 0x0000
)

func (f flagsType) isValid() bool {
	return f == flagsZero
}

func (a acceptType) isValid() bool {
	return a == acceptType(0x00) //nolint:gomnd // Only to be set on sent headers
}

const (
	contentTypeProtobuf contentType = 0x00
)

func (c contentType) isValid() bool {
	return c == contentTypeProtobuf
}

// magic number to indicate we have a wire header
const magicNumber uint32 = 0x5EC0A710

// wireHeader represents a request header
type wireHeader struct {
	magicNumber  uint32
	hdrSize      uint16
	versionMajor versionMajorType
	versionMinor versionMinorType
	flags        flagsType
	provider     ProviderID
	session      sessionType
	contentType  contentType
	acceptType   acceptType
	authType     auth.AuthenticationType
	bodyLen      uint32
	authLen      uint16
	opCode       OpCode
	Status       StatusCode
	reserved1    uint8
	reserved2    uint8
}

func (r *wireHeader) pack(buf *bytes.Buffer) error {
	// panic rather than error as this is internal api and this shouldn't happen
	if buf == nil {
		panic("buffer pointer is nil")
	}

	// Check values are correct before packing

	r.magicNumber = magicNumber
	r.hdrSize = requestHeaderSize

	if err := r.checkForRequest(); err != nil {
		return err
	}

	err := binary.Write(buf, binary.LittleEndian, r)
	return err
}

func (r *wireHeader) checkForRequest() error {
	if !isSupportedWireHeaderVersion(r.versionMajor, r.versionMinor) {
		return fmt.Errorf("invalid version %v.%v", r.versionMajor, r.versionMinor)
	}
	if !r.flags.isValid() {
		return fmt.Errorf("invalid flags %v", r.flags)
	}
	if !r.contentType.isValid() {
		return fmt.Errorf("invalid content type %v", r.contentType)
	}
	if !r.acceptType.isValid() {
		return fmt.Errorf("invalid accept type %v", r.acceptType)
	}
	if !r.authType.IsValid() {
		return fmt.Errorf("invaliid auth type %v", r.authType)
	}
	return nil
}

func isSupportedWireHeaderVersion(maj versionMajorType, min versionMinorType) bool {
	return maj == versionMajorOne && min == versionMinorZero
}

const (
	buffBytes8Bit  int = 1
	buffBytes16Bit int = 2
	buffBytes32Bit int = 4
	buffBytes64Bit int = 8
)

func parseWireHeaderFromBuf(buf *bytes.Buffer) (*wireHeader, error) {
	r := &wireHeader{}
	// panic rather than error as this is internal api and this shouldn't happen
	if buf == nil {
		panic("buffer pointer is nil")
	}
	r.magicNumber = binary.LittleEndian.Uint32(buf.Next(buffBytes32Bit))
	if r.magicNumber != magicNumber {
		return nil, fmt.Errorf("invalid magic number")
	}
	r.hdrSize = binary.LittleEndian.Uint16(buf.Next(buffBytes16Bit))
	if r.hdrSize != wireHeaderSizeValue {
		return nil, fmt.Errorf("invalid header size (%d != %d)", r.hdrSize, wireHeaderSizeValue)
	}
	r.versionMajor = versionMajorType(buf.Next(buffBytes8Bit)[0])
	r.versionMinor = versionMinorType(buf.Next(buffBytes8Bit)[0])
	if !isSupportedWireHeaderVersion(r.versionMajor, r.versionMinor) {
		return nil, fmt.Errorf("unsupported version number %v.%v", r.versionMajor, r.versionMinor)
	}
	r.flags = flagsType(binary.LittleEndian.Uint16(buf.Next(buffBytes16Bit)))
	if !r.flags.isValid() {
		return nil, fmt.Errorf("unsupported flags value %v", r.flags)
	}
	r.provider = ProviderID(buf.Next(buffBytes8Bit)[0])
	if !r.provider.IsValid() {
		return nil, fmt.Errorf("invalid provider %v", r.provider)
	}
	r.session = sessionType(binary.LittleEndian.Uint64(buf.Next(buffBytes64Bit))) // Can take any value in range
	r.contentType = contentType(buf.Next(buffBytes8Bit)[0])
	if !r.contentType.isValid() {
		return nil, fmt.Errorf("invalid content type %v", r.contentType)
	}
	r.acceptType = acceptType(buf.Next(buffBytes8Bit)[0]) // This should only be set in requests so we must not check value
	r.authType = auth.AuthenticationType(buf.Next(buffBytes8Bit)[0])
	if !r.authType.IsValid() {
		return nil, fmt.Errorf("invalid auth type %v", r.authType)
	}
	r.bodyLen = binary.LittleEndian.Uint32(buf.Next(buffBytes32Bit))
	r.authLen = binary.LittleEndian.Uint16(buf.Next(buffBytes16Bit))
	r.opCode = OpCode(binary.LittleEndian.Uint32(buf.Next(buffBytes32Bit)))
	if !r.opCode.IsValid() {
		return nil, fmt.Errorf("invalid opcode %v", r.opCode)
	}
	r.Status = StatusCode(binary.LittleEndian.Uint16(buf.Next(buffBytes16Bit)))
	if !r.Status.IsValid() {
		return nil, fmt.Errorf("invalid response status code %v", r.Status)
	}
	r.reserved1 = buf.Next(buffBytes8Bit)[0]
	r.reserved2 = buf.Next(buffBytes8Bit)[0]
	if r.reserved1 != 0x00 || r.reserved2 != 0x00 {
		return nil, fmt.Errorf("reserved bytes must be zero")
	}

	return r, nil
}
