//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package requests

import (
	"encoding/json"
	"os"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/fxamacker/cbor/v2"
)

// AddEventRequest defines the Request Content for POST event DTO.
type AddEventRequest struct {
	dtoCommon.BaseRequest `json:",inline"`
	Event                 dtos.Event `json:"event" validate:"required"`
}

// NewAddEventRequest creates, initializes and returns an AddEventRequests
func NewAddEventRequest(event dtos.Event) AddEventRequest {
	return AddEventRequest{
		BaseRequest: dtoCommon.NewBaseRequest(),
		Event:       event,
	}
}

// Validate satisfies the Validator interface
func (a AddEventRequest) Validate() error {
	if err := common.Validate(a); err != nil {
		return err
	}

	// BaseReading has the skip("-") validation annotation for BinaryReading and SimpleReading
	// Otherwise error will occur as only one of them exists
	// Therefore, need to validate the nested BinaryReading and SimpleReading struct here
	for _, r := range a.Event.Readings {
		if err := r.Validate(); err != nil {
			return err
		}
	}

	return nil
}

type unmarshal func([]byte, interface{}) error

func (a *AddEventRequest) UnmarshalJSON(b []byte) error {
	return a.Unmarshal(b, json.Unmarshal)
}

func (a *AddEventRequest) UnmarshalCBOR(b []byte) error {
	return a.Unmarshal(b, cbor.Unmarshal)
}

func (a *AddEventRequest) Unmarshal(b []byte, f unmarshal) error {
	// To avoid recursively invoke unmarshaler interface, intentionally create a struct to represent AddEventRequest DTO
	var addEvent struct {
		dtoCommon.BaseRequest
		Event dtos.Event
	}
	if err := f(b, &addEvent); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal the byte array.", err)
	}

	*a = AddEventRequest(addEvent)

	// validate AddEventRequest DTO
	if err := a.Validate(); err != nil {
		return err
	}

	// Normalize reading's value type
	for i, r := range a.Event.Readings {
		valueType, err := common.NormalizeValueType(r.ValueType)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
		a.Event.Readings[i].ValueType = valueType
	}
	return nil
}

func (a *AddEventRequest) Encode() ([]byte, string, error) {
	var encoding = a.GetEncodingContentType()
	var err error
	var encodedData []byte

	switch encoding {
	case common.ContentTypeCBOR:
		encodedData, err = cbor.Marshal(a)
		if err != nil {
			return nil, "", errors.NewCommonEdgeX(errors.KindContractInvalid, "failed to encode AddEventRequest to CBOR", err)
		}
	case common.ContentTypeJSON:
		encodedData, err = json.Marshal(a)
		if err != nil {
			return nil, "", errors.NewCommonEdgeX(errors.KindContractInvalid, "failed to encode AddEventRequest to JSON", err)
		}
	}

	return encodedData, encoding, nil
}

// GetEncodingContentType determines which content type should be used to encode and decode this object
func (a *AddEventRequest) GetEncodingContentType() string {
	if v := os.Getenv(common.EnvEncodeAllEvents); v == common.ValueTrue {
		return common.ContentTypeCBOR
	}
	for _, r := range a.Event.Readings {
		if r.ValueType == common.ValueTypeBinary {
			return common.ContentTypeCBOR
		}
	}

	return common.ContentTypeJSON
}

// AddEventReqToEventModel transforms the AddEventRequest DTO to the Event model
func AddEventReqToEventModel(addEventReq AddEventRequest) (event models.Event) {
	readings := make([]models.Reading, len(addEventReq.Event.Readings))
	for i, r := range addEventReq.Event.Readings {
		readings[i] = dtos.ToReadingModel(r)
	}

	tags := make(map[string]interface{})
	for tag, value := range addEventReq.Event.Tags {
		tags[tag] = value
	}

	return models.Event{
		Id:          addEventReq.Event.Id,
		DeviceName:  addEventReq.Event.DeviceName,
		ProfileName: addEventReq.Event.ProfileName,
		SourceName:  addEventReq.Event.SourceName,
		Origin:      addEventReq.Event.Origin,
		Readings:    readings,
		Tags:        tags,
	}
}
