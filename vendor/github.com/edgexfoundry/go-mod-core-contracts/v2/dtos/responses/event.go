//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package responses

import (
	"encoding/json"
	"os"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/fxamacker/cbor/v2"
)

// EventResponse defines the Response Content for GET event DTOs.
// This object and its properties correspond to the EventResponse object in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-data/2.1.0#/EventResponse
type EventResponse struct {
	dtoCommon.BaseResponse `json:",inline"`
	Event                  dtos.Event `json:"event"`
}

// MultiEventsResponse defines the Response Content for GET multiple event DTOs.
// This object and its properties correspond to the MultiEventsResponse object in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-data/2.1.0#/MultiEventsResponse
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
	var encoding = common.ContentTypeJSON

	for _, r := range e.Event.Readings {
		if r.ValueType == common.ValueTypeBinary {
			encoding = common.ContentTypeCBOR
			break
		}
	}
	if v := os.Getenv(common.EnvEncodeAllEvents); v == common.ValueTrue {
		encoding = common.ContentTypeCBOR
	}

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
