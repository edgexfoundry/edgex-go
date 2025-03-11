//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package responses

import (
	"encoding/json"
	"os"

	"github.com/fxamacker/cbor/v2"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

// EventResponse defines the Response Content for GET event DTOs.
type EventResponse struct {
	dtoCommon.BaseResponse `json:",inline"`
	Event                  dtos.Event `json:"event"`
}

// MultiEventsResponse defines the Response Content for GET multiple event DTOs.
type MultiEventsResponse struct {
	dtoCommon.BaseWithTotalCountResponse `json:",inline"`
	Events                               []dtos.Event `json:"events"`
}

func NewEventResponse(requestId string, message string, statusCode int, event dtos.Event) EventResponse {
	return EventResponse{
		BaseResponse: dtoCommon.NewBaseResponse(requestId, message, statusCode),
		Event:        event,
	}
}

func NewMultiEventsResponse(requestId string, message string, statusCode int, totalCount uint32, events []dtos.Event) MultiEventsResponse {
	return MultiEventsResponse{
		BaseWithTotalCountResponse: dtoCommon.NewBaseWithTotalCountResponse(requestId, message, statusCode, totalCount),
		Events:                     events,
	}
}

func (e *EventResponse) Encode() ([]byte, string, error) {
	var encoding = e.GetEncodingContentType()
	var err error
	var encodedData []byte

	switch encoding {
	case common.ContentTypeCBOR:
		encodedData, err = cbor.Marshal(e)
		if err != nil {
			return nil, "", errors.NewCommonEdgeX(errors.KindContractInvalid, "failed to encode EventResponse to CBOR", err)
		}
	case common.ContentTypeJSON:
		encodedData, err = json.Marshal(e)
		if err != nil {
			return nil, "", errors.NewCommonEdgeX(errors.KindContractInvalid, "failed to encode EventResponse to JSON", err)
		}
	}

	return encodedData, encoding, nil
}

// GetEncodingContentType determines which content type should be used to encode and decode this object
func (e *EventResponse) GetEncodingContentType() string {
	if v := os.Getenv(common.EnvEncodeAllEvents); v == common.ValueTrue {
		return common.ContentTypeCBOR
	}
	for _, r := range e.Event.Readings {
		if r.ValueType == common.ValueTypeBinary {
			return common.ContentTypeCBOR
		}
	}

	return common.ContentTypeJSON
}
