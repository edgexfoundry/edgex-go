//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package io

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"

	"github.com/fxamacker/cbor/v2"
	"gopkg.in/yaml.v3"
)

type DtoReader interface {
	Read(reader io.Reader, v interface{}) errors.EdgeX
}

// NewDtoReader returns a BodyReader capable of processing the request body
func NewDtoReader(contentType string) DtoReader {
	switch strings.ToLower(contentType) {
	case common.ContentTypeCBOR:
		return NewCborDtoReader()
	default:
		return NewJsonDtoReader()
	}
}

type jsonDtoReader struct{}

func NewJsonDtoReader() DtoReader {
	return jsonDtoReader{}
}

// Read reads JSON data and decodes it into the specified DTO
func (jsonDtoReader) Read(reader io.Reader, v interface{}) errors.EdgeX {
	err := json.NewDecoder(reader).Decode(v)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	return nil
}

type cborDtoReader struct{}

func NewCborDtoReader() DtoReader {
	return cborDtoReader{}
}

// Read reads CBOR data and decodes it into the specified DTO
func (cborDtoReader) Read(reader io.Reader, v interface{}) errors.EdgeX {
	err := cbor.NewDecoder(reader).Decode(v)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("%T cbor decoding failed", v), err)
	}

	return nil
}

type yamlDtoReader struct{}

func NewYamlDtoReader() DtoReader {
	return yamlDtoReader{}
}

// Read reads Yaml data and decodes it into the specified DTO
func (yamlDtoReader) Read(reader io.Reader, v interface{}) errors.EdgeX {
	err := yaml.NewDecoder(reader).Decode(v)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("%T yaml decoding failed", v), err)
	}

	return nil
}
