//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package responses

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
)

// ScheduleActionRecordResponse defines the Response Content for GET ScheduleActionRecord DTO.
type ScheduleActionRecordResponse struct {
	common.BaseResponse  `json:",inline"`
	ScheduleActionRecord dtos.ScheduleActionRecord `json:"scheduleActionRecord"`
}

func NewScheduleActionRecordResponse(requestId string, message string, statusCode int, scheduleActionRecord dtos.ScheduleActionRecord) ScheduleActionRecordResponse {
	return ScheduleActionRecordResponse{
		BaseResponse:         common.NewBaseResponse(requestId, message, statusCode),
		ScheduleActionRecord: scheduleActionRecord,
	}
}

// MultiScheduleActionRecordsResponse defines the Response Content for GET multiple ScheduleActionRecord DTOs.
type MultiScheduleActionRecordsResponse struct {
	common.BaseWithTotalCountResponse `json:",inline"`
	ScheduleActionRecords             []dtos.ScheduleActionRecord `json:"scheduleActionRecords"`
}

func NewMultiScheduleActionRecordsResponse(requestId string, message string, statusCode int, totalCount uint32, scheduleActionRecords []dtos.ScheduleActionRecord) MultiScheduleActionRecordsResponse {
	return MultiScheduleActionRecordsResponse{
		BaseWithTotalCountResponse: common.NewBaseWithTotalCountResponse(requestId, message, statusCode, totalCount),
		ScheduleActionRecords:      scheduleActionRecords,
	}
}
