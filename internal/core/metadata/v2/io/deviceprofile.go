//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package io

import (
	"encoding/json"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"

	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	dto "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
)

// DeviceProfileReader unmarshals a request body into an DeviceProfile type
type DeviceProfileReader interface {
	ReadDeviceProfileRequest(reader io.Reader) ([]dto.DeviceProfileRequest, errors.EdgeX)
	ReadDeviceProfileYaml(r *http.Request) (dtos.DeviceProfile, errors.EdgeX)
}

// NewRequestReader returns a BodyReader capable of processing the request body
func NewDeviceProfileRequestReader() DeviceProfileReader {
	return NewJsonReader()
}

// jsonReader handles unmarshaling of a JSON request body payload
type jsonDeviceProfileReader struct{}

// NewJsonReader creates a new instance of jsonReader.
func NewJsonReader() jsonDeviceProfileReader {
	return jsonDeviceProfileReader{}
}

// ReadDeviceProfileRequest reads and converts the request's JSON data into an DeviceProfile struct
func (jsonDeviceProfileReader) ReadDeviceProfileRequest(reader io.Reader) ([]dto.DeviceProfileRequest, errors.EdgeX) {
	var addDeviceProfiles []dto.DeviceProfileRequest
	err := json.NewDecoder(reader).Decode(&addDeviceProfiles)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "device profile json decoding failed", err)
	}
	return addDeviceProfiles, nil
}

// ReadDeviceProfileYaml reads and converts the request's YAML file into an DeviceProfile struct
func (jsonDeviceProfileReader) ReadDeviceProfileYaml(r *http.Request) (dtos.DeviceProfile, errors.EdgeX) {
	var f multipart.File
	f, _, err := r.FormFile("file")
	if err != nil {
		return dtos.DeviceProfile{}, errors.NewCommonEdgeX(errors.KindContractInvalid, "missing yaml file", err)
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return dtos.DeviceProfile{}, errors.NewCommonEdgeX(errors.KindServerError, "failed to read yaml file", err)
	}
	if len(data) == 0 {
		return dtos.DeviceProfile{}, errors.NewCommonEdgeX(errors.KindContractInvalid, "yaml file is empty", err)
	}

	var dp dtos.DeviceProfile

	err = yaml.Unmarshal(data, &dp)
	if err != nil {
		return dtos.DeviceProfile{}, errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to unmarshal yaml file", err)
	}
	err = v2.Validate(dp)
	if err != nil {
		return dtos.DeviceProfile{}, errors.NewCommonEdgeXWrapper(err)
	}

	return dp, nil
}
