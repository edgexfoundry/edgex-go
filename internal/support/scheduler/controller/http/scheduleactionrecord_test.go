//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/container"
	csMock "github.com/edgexfoundry/edgex-go/internal/support/scheduler/infrastructure/interfaces/mocks"
)

func scheduleActionRecordsData() []dtos.ScheduleActionRecord {
	return []dtos.ScheduleActionRecord{
		{
			Id:          exampleUUID,
			JobName:     testScheduleJobName,
			Action:      testScheduleAction,
			Status:      testStatus,
			ScheduledAt: testTimestamp,
			Created:     testTimestamp,
		},
		{
			Id:          "2e2682eb-0f24-48aa-ae4c-de9dac3fb9bc",
			JobName:     "jobName2",
			Action:      testScheduleAction,
			Status:      testStatus,
			ScheduledAt: testTimestamp,
			Created:     testTimestamp,
		},
	}
}

func TestAllScheduleActionRecords(t *testing.T) {
	expectedTotalScheduleActionRecordCount := int64(2)
	dic := mockDic()
	dbClientMock := &csMock.DBClient{}
	dbClientMock.On("ScheduleActionRecordTotalCount", context.Background(), int64(0), mock.AnythingOfType("int64")).Return(expectedTotalScheduleActionRecordCount, nil)
	dbClientMock.On("AllScheduleActionRecords", context.Background(), int64(0), mock.AnythingOfType("int64"), 0, 20).Return([]models.ScheduleActionRecord{}, nil)
	dbClientMock.On("AllScheduleActionRecords", context.Background(), int64(0), mock.AnythingOfType("int64"), 0, 1).Return([]models.ScheduleActionRecord{}, nil)
	dbClientMock.On("AllScheduleActionRecords", context.Background(), int64(1723642430000), int64(1723642440000), 0, 1).Return([]models.ScheduleActionRecord{}, nil)
	dbClientMock.On("ScheduleActionRecordTotalCount", context.Background(), int64(1723642430000), int64(1723642440000)).Return(expectedTotalScheduleActionRecordCount, nil)
	dbClientMock.On("AllScheduleActionRecords", context.Background(), int64(0), mock.AnythingOfType("int64"), 4, 2).Return([]models.ScheduleActionRecord{}, errors.NewCommonEdgeX(errors.KindRangeNotSatisfiable, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) any {
			return dbClientMock
		},
	})
	controller := NewScheduleActionRecordController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		start              string
		end                string
		offset             string
		limit              string
		errorExpected      bool
		expectedTotalCount int64
		expectedStatusCode int
	}{
		{"Valid - get schedule action records without start, end, offset, and limit", "", "", "", "", false, expectedTotalScheduleActionRecordCount, http.StatusOK},
		{"Valid - get schedule action records without start and end and with offset and limit", "", "", "0", "1", false, expectedTotalScheduleActionRecordCount, http.StatusOK},
		{"Valid - get schedule action records without start and with end, offset, and limit", "", "1723642440000", "0", "1", false, expectedTotalScheduleActionRecordCount, http.StatusOK},
		{"Valid - get schedule action records with start, end, offset, and limit", "1723642430000", "1723642440000", "0", "1", false, expectedTotalScheduleActionRecordCount, http.StatusOK},
		{"Invalid - invalid start, end must be greater than start", "1723642440000", "0", "", "", true, expectedTotalScheduleActionRecordCount, http.StatusBadRequest},
		{"Invalid - invalid offset format", "", "", "aaa", "1", true, expectedTotalScheduleActionRecordCount, http.StatusBadRequest},
		{"Invalid - invalid limit format", "", "", "1", "aaa", true, expectedTotalScheduleActionRecordCount, http.StatusBadRequest},
		{"Invalid - offset out of range", "", "", "4", "2", true, expectedTotalScheduleActionRecordCount, http.StatusRequestedRangeNotSatisfiable},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			e := echo.New()
			req, err := http.NewRequest(http.MethodGet, common.ApiAllScheduleActionRecordRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.start != "" {
				query.Add(common.Start, testCase.start)
			}
			if testCase.end != "" {
				query.Add(common.End, testCase.end)
			}
			if testCase.offset != "" {
				query.Add(common.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(common.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			c := e.NewContext(req, recorder)
			err = controller.AllScheduleActionRecords(c)
			require.NoError(t, err)

			// Assert
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiScheduleActionRecordsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Response total count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestLatestScheduleActionRecordsByJobName(t *testing.T) {
	expectedTotalScheduleActionRecordCount := int64(0)
	emptyJobName := ""
	notFoundJobName := "notFoundJobName"
	dic := mockDic()
	dbClientMock := &csMock.DBClient{}
	dbClientMock.On("ScheduleJobByName", context.Background(), testScheduleJobName).Return(models.ScheduleJob{}, nil)
	dbClientMock.On("LatestScheduleActionRecordsByJobName", context.Background(), testScheduleJobName).Return([]models.ScheduleActionRecord{}, nil)
	dbClientMock.On("ScheduleJobByName", context.Background(), emptyJobName).Return(models.ScheduleJob{}, errors.NewCommonEdgeX(errors.KindContractInvalid, "scheduled job doesn't exist in the database", nil))
	dbClientMock.On("LatestScheduleActionRecordsByJobName", context.Background(), emptyJobName).Return([]models.ScheduleActionRecord{}, nil)
	dbClientMock.On("ScheduleJobByName", context.Background(), notFoundJobName).Return(models.ScheduleJob{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "scheduled job doesn't exist in the database", nil))
	dbClientMock.On("LatestScheduleActionRecordsByJobName", context.Background(), notFoundJobName).Return([]models.ScheduleActionRecord{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) any {
			return dbClientMock
		},
	})
	controller := NewScheduleActionRecordController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		jobName            string
		errorExpected      bool
		expectedTotalCount int64
		expectedStatusCode int
	}{
		{"Valid - get schedule action records with offset and limit", testScheduleJobName, false, expectedTotalScheduleActionRecordCount, http.StatusOK},
		{"Invalid - invalid empty jobName", emptyJobName, true, expectedTotalScheduleActionRecordCount, http.StatusBadRequest},
		{"Invalid - job not found by name", notFoundJobName, true, expectedTotalScheduleActionRecordCount, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			e := echo.New()
			reqPath := fmt.Sprintf("%s/%s", common.ApiLatestScheduleActionRecordByJobNameRoute, testCase.jobName)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			c := e.NewContext(req, recorder)
			c.SetParamNames(common.Name)
			c.SetParamValues(testCase.jobName)
			err = controller.LatestScheduleActionRecordsByJobName(c)
			require.NoError(t, err)

			// Assert
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiScheduleActionRecordsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Equal(t, testCase.expectedTotalCount, res.TotalCount, "Response total count not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestScheduleActionRecordsByStatus(t *testing.T) {
	expectedTotalScheduleActionRecordCount := int64(2)

	var records []models.ScheduleActionRecord
	for _, dto := range scheduleActionRecordsData() {
		records = append(records, dtos.ToScheduleActionRecordModel(dto))
	}

	emptyStatus := ""
	notFoundStatus := "notFoundStatus"

	dic := mockDic()
	dbClientMock := &csMock.DBClient{}
	dbClientMock.On("ScheduleActionRecordCountByStatus", context.Background(), testStatus, int64(0), mock.AnythingOfType("int64")).Return(expectedTotalScheduleActionRecordCount, nil)
	dbClientMock.On("ScheduleActionRecordsByStatus", context.Background(), testStatus, int64(0), mock.AnythingOfType("int64"), 0, 20).Return(records, nil)
	dbClientMock.On("ScheduleActionRecordCountByStatus", context.Background(), notFoundStatus, int64(0), mock.AnythingOfType("int64")).Return(expectedTotalScheduleActionRecordCount, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "schedule action records with given status doesn't exist in the database", nil))
	dbClientMock.On("ScheduleActionRecordsByStatus", context.Background(), testStatus, int64(0), mock.AnythingOfType("int64"), 4, 2).Return([]models.ScheduleActionRecord{}, errors.NewCommonEdgeX(errors.KindRangeNotSatisfiable, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) any {
			return dbClientMock
		},
	})

	controller := NewScheduleActionRecordController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		status             string
		start              string
		end                string
		offset             string
		limit              string
		errorExpected      bool
		expectedTotalCount int64
		expectedStatusCode int
	}{
		{"Valid - find schedule action records by status", testStatus, "", "", "", "", false, expectedTotalScheduleActionRecordCount, http.StatusOK},
		{"Invalid - status parameter is empty", emptyStatus, "", "", "", "", true, expectedTotalScheduleActionRecordCount, http.StatusBadRequest},
		{"Invalid - schedule action records not found by status", notFoundStatus, "", "", "", "", true, expectedTotalScheduleActionRecordCount, http.StatusNotFound},
		{"Invalid - offset out of range", testStatus, "", "", "4", "2", true, expectedTotalScheduleActionRecordCount, http.StatusRequestedRangeNotSatisfiable},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			e := echo.New()
			reqPath := fmt.Sprintf("%s/%s", common.ApiScheduleActionRecordRouteByStatusRoute, testCase.status)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			query := req.URL.Query()
			if testCase.start != "" {
				query.Add(common.Start, testCase.start)
			}
			if testCase.end != "" {
				query.Add(common.End, testCase.end)
			}
			if testCase.offset != "" {
				query.Add(common.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(common.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			c := e.NewContext(req, recorder)
			c.SetParamNames(common.Status)
			c.SetParamValues(testCase.status)
			err = controller.ScheduleActionRecordsByStatus(c)
			require.NoError(t, err)

			// Assert
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiScheduleActionRecordsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Equal(t, testCase.status, res.ScheduleActionRecords[0].Status, "Status not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestScheduleActionRecordsByJobName(t *testing.T) {
	expectedTotalScheduleActionRecordCount := int64(2)

	var records []models.ScheduleActionRecord
	for _, dto := range scheduleActionRecordsData() {
		records = append(records, dtos.ToScheduleActionRecordModel(dto))
	}

	emptyJobName := ""
	notFoundJobName := "notFoundJobName"

	dic := mockDic()
	dbClientMock := &csMock.DBClient{}
	dbClientMock.On("ScheduleActionRecordCountByJobName", context.Background(), testScheduleJobName, int64(0), mock.AnythingOfType("int64")).Return(expectedTotalScheduleActionRecordCount, nil)
	dbClientMock.On("ScheduleActionRecordsByJobName", context.Background(), testScheduleJobName, int64(0), mock.AnythingOfType("int64"), 0, 20).Return(records, nil)
	dbClientMock.On("ScheduleActionRecordCountByJobName", context.Background(), notFoundJobName, int64(0), mock.AnythingOfType("int64")).Return(expectedTotalScheduleActionRecordCount, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "schedule action records with given job name doesn't exist in the database", nil))
	dbClientMock.On("ScheduleActionRecordsByJobName", context.Background(), testScheduleJobName, int64(0), mock.AnythingOfType("int64"), 4, 2).Return([]models.ScheduleActionRecord{}, errors.NewCommonEdgeX(errors.KindRangeNotSatisfiable, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) any {
			return dbClientMock
		},
	})

	controller := NewScheduleActionRecordController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		jobName            string
		start              string
		end                string
		offset             string
		limit              string
		errorExpected      bool
		expectedTotalCount int64
		expectedStatusCode int
	}{
		{"Valid - find schedule action records by job name", testScheduleJobName, "", "", "", "", false, expectedTotalScheduleActionRecordCount, http.StatusOK},
		{"Invalid - job name parameter is empty", emptyJobName, "", "", "", "", true, expectedTotalScheduleActionRecordCount, http.StatusBadRequest},
		{"Invalid - schedule action records not found by job name", notFoundJobName, "", "", "", "", true, expectedTotalScheduleActionRecordCount, http.StatusNotFound},
		{"Invalid - offset out of range", testScheduleJobName, "", "", "4", "2", true, expectedTotalScheduleActionRecordCount, http.StatusRequestedRangeNotSatisfiable},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			e := echo.New()
			reqPath := fmt.Sprintf("%s/%s", common.ApiScheduleActionRecordRouteByJobNameRoute, testCase.jobName)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			query := req.URL.Query()
			if testCase.start != "" {
				query.Add(common.Start, testCase.start)
			}
			if testCase.end != "" {
				query.Add(common.End, testCase.end)
			}
			if testCase.offset != "" {
				query.Add(common.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(common.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			c := e.NewContext(req, recorder)
			c.SetParamNames(common.Name)
			c.SetParamValues(testCase.jobName)
			err = controller.ScheduleActionRecordsByJobName(c)
			require.NoError(t, err)

			// Assert
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiScheduleActionRecordsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Equal(t, testCase.jobName, res.ScheduleActionRecords[0].JobName, "JobName not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestScheduleActionRecordsByJobNameAndStatus(t *testing.T) {
	expectedTotalScheduleActionRecordCount := int64(2)

	var records []models.ScheduleActionRecord
	for _, dto := range scheduleActionRecordsData() {
		records = append(records, dtos.ToScheduleActionRecordModel(dto))
	}

	emptyJobName := ""
	emptyStatus := ""
	notFoundJobName := "notFoundJobName"
	notFoundStatus := "notFoundStatus"

	dic := mockDic()
	dbClientMock := &csMock.DBClient{}
	dbClientMock.On("ScheduleActionRecordCountByJobNameAndStatus", context.Background(), testScheduleJobName, testStatus, int64(0), mock.AnythingOfType("int64")).Return(expectedTotalScheduleActionRecordCount, nil)
	dbClientMock.On("ScheduleActionRecordsByJobNameAndStatus", context.Background(), testScheduleJobName, testStatus, int64(0), mock.AnythingOfType("int64"), 0, 20).Return(records, nil)
	dbClientMock.On("ScheduleActionRecordCountByJobNameAndStatus", context.Background(), notFoundJobName, notFoundStatus, int64(0), mock.AnythingOfType("int64")).Return(expectedTotalScheduleActionRecordCount, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "schedule action records with given job name and status doesn't exist in the database", nil))
	dbClientMock.On("ScheduleActionRecordsByJobNameAndStatus", context.Background(), testScheduleJobName, testStatus, int64(0), mock.AnythingOfType("int64"), 4, 2).Return([]models.ScheduleActionRecord{}, errors.NewCommonEdgeX(errors.KindRangeNotSatisfiable, "query objects bounds out of range.", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) any {
			return dbClientMock
		},
	})

	controller := NewScheduleActionRecordController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		jobName            string
		status             string
		start              string
		end                string
		offset             string
		limit              string
		errorExpected      bool
		expectedTotalCount int64
		expectedStatusCode int
	}{
		{"Valid - find schedule action records by job name", testScheduleJobName, testStatus, "", "", "", "", false, expectedTotalScheduleActionRecordCount, http.StatusOK},
		{"Invalid - job name and status parameters are empty", emptyJobName, emptyStatus, "", "", "", "", true, expectedTotalScheduleActionRecordCount, http.StatusBadRequest},
		{"Invalid - schedule action records not found by job name and status", notFoundJobName, notFoundStatus, "", "", "", "", true, expectedTotalScheduleActionRecordCount, http.StatusNotFound},
		{"Invalid - offset out of range", testScheduleJobName, testStatus, "", "", "4", "2", true, expectedTotalScheduleActionRecordCount, http.StatusRequestedRangeNotSatisfiable},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			e := echo.New()
			reqPath := fmt.Sprintf("%s/%s/%s/:%s/%s", common.ApiScheduleActionRecordRouteByJobNameRoute, testCase.jobName, common.Status, common.Status, testCase.status)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			query := req.URL.Query()
			if testCase.start != "" {
				query.Add(common.Start, testCase.start)
			}
			if testCase.end != "" {
				query.Add(common.End, testCase.end)
			}
			if testCase.offset != "" {
				query.Add(common.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(common.Limit, testCase.limit)
			}
			req.URL.RawQuery = query.Encode()
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			c := e.NewContext(req, recorder)
			c.SetParamNames(common.Name, common.Status)
			c.SetParamValues(testCase.jobName, testCase.status)

			err = controller.ScheduleActionRecordsByJobNameAndStatus(c)
			require.NoError(t, err)

			// Assert
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.MultiScheduleActionRecordsResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
				assert.Equal(t, testCase.jobName, res.ScheduleActionRecords[0].JobName, "JobName not as expected")
				assert.Equal(t, testCase.status, res.ScheduleActionRecords[0].Status, "Status not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}
