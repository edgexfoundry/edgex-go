//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v4/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/config"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/container"
	csMock "github.com/edgexfoundry/edgex-go/internal/support/scheduler/infrastructure/interfaces/mocks"
)

func mockDic() *di.Container {
	return di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) any {
			return &config.ConfigurationStruct{
				Writable: config.WritableInfo{
					LogLevel: "DEBUG",
				},
				Service: bootstrapConfig.ServiceInfo{
					MaxResultCount: 20,
				},
			}
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) any {
			return logger.NewMockClient()
		},
	})
}

func addScheduleJobRequestData() requests.AddScheduleJobRequest {
	return requests.AddScheduleJobRequest{
		BaseRequest: commonDTO.BaseRequest{
			RequestId:   exampleUUID,
			Versionable: commonDTO.NewVersionable(),
		},
		ScheduleJob: dtos.ScheduleJob{
			Name:       testScheduleJobName,
			Definition: testScheduleDef,
			Actions:    testScheduleActions,
			Labels:     testScheduleJobLabels,
		},
	}
}

func updateScheduleJobRequestData() requests.UpdateScheduleJobRequest {
	id := exampleUUID
	name := testScheduleJobName
	definition := testScheduleDef
	actions := testScheduleActions
	labels := testScheduleJobLabels

	var req = requests.UpdateScheduleJobRequest{
		BaseRequest: commonDTO.BaseRequest{
			RequestId:   exampleUUID,
			Versionable: commonDTO.NewVersionable(),
		},
		ScheduleJob: dtos.UpdateScheduleJob{
			Id:         &id,
			Name:       &name,
			Definition: &definition,
			Actions:    actions,
			Labels:     labels,
		},
	}

	return req
}

func TestAddScheduleJob(t *testing.T) {
	expectedRequestId := exampleUUID
	dic := mockDic()
	dbClientMock := &csMock.DBClient{}
	schedulerManagerMock := &csMock.SchedulerManager{}

	valid := addScheduleJobRequestData()
	model := dtos.ToScheduleJobModel(valid.ScheduleJob)
	dbClientMock.On("AddScheduleJob", context.Background(), mock.MatchedBy(func(job models.ScheduleJob) bool { return job.Name == testScheduleJobName })).Return(model, nil)
	schedulerManagerMock.On("AddScheduleJob", mock.MatchedBy(func(job models.ScheduleJob) bool { return job.Name == testScheduleJobName }), testCorrelationID).Return(nil)
	schedulerManagerMock.On("StartScheduleJobByName", model.Name, testCorrelationID).Return(nil)

	noName := addScheduleJobRequestData()
	noName.ScheduleJob.Name = ""
	noRequestId := addScheduleJobRequestData()
	noRequestId.RequestId = ""

	duplicatedName := addScheduleJobRequestData()
	duplicatedName.ScheduleJob.Name = "duplicatedName"
	model = dtos.ToScheduleJobModel(duplicatedName.ScheduleJob)
	schedulerManagerMock.On(
		"AddScheduleJob",
		mock.MatchedBy(
			func(job models.ScheduleJob) bool {
				return job.Name == "duplicatedName"
			}),
		testCorrelationID,
	).Return(errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("scheduled job name %s already exists", model.Name), nil))

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) any {
			return dbClientMock
		},
		container.SchedulerManagerName: func(get di.Get) any {
			return schedulerManagerMock
		},
	})
	controller := NewScheduleJobController(dic)
	assert.NotNil(t, controller)
	tests := []struct {
		name               string
		request            []requests.AddScheduleJobRequest
		expectedStatusCode int
	}{
		{"Valid", []requests.AddScheduleJobRequest{valid}, http.StatusCreated},
		{"Valid - no request Id", []requests.AddScheduleJobRequest{noRequestId}, http.StatusCreated},
		{"Invalid - no name", []requests.AddScheduleJobRequest{noName}, http.StatusBadRequest},
		{"Invalid - duplicated name", []requests.AddScheduleJobRequest{duplicatedName}, http.StatusConflict},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			e := echo.New()
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPost, common.ApiScheduleJobRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			c := e.NewContext(req, recorder)
			err = controller.AddScheduleJob(c)
			require.NoError(t, err)
			if testCase.expectedStatusCode == http.StatusBadRequest {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "BaseResponse status code not as expected")
				assert.NotEmpty(t, res.Message, "Message is empty")
			} else {
				var res []commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, http.StatusMultiStatus, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, common.ApiVersion, res[0].ApiVersion, "API Version not as expected")
				if res[0].RequestId != "" {
					assert.Equal(t, expectedRequestId, res[0].RequestId, "RequestID not as expected")
				}
				assert.Equal(t, testCase.expectedStatusCode, res[0].StatusCode, "BaseResponse status code not as expected")
			}
		})
	}
}

func TestScheduleJobByName(t *testing.T) {
	job := dtos.ToScheduleJobModel(addScheduleJobRequestData().ScheduleJob)
	emptyName := ""
	notFoundName := "notFoundName"

	dic := mockDic()
	dbClientMock := &csMock.DBClient{}
	dbClientMock.On("ScheduleJobByName", context.Background(), job.Name).Return(job, nil)
	dbClientMock.On("ScheduleJobByName", context.Background(), notFoundName).Return(models.ScheduleJob{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "scheduled job doesn't exist in the database", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) any {
			return dbClientMock
		},
	})

	controller := NewScheduleJobController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		jobName            string
		errorExpected      bool
		expectedStatusCode int
	}{
		{"Valid - find scheduled job by name", job.Name, false, http.StatusOK},
		{"Invalid - name parameter is empty", emptyName, true, http.StatusBadRequest},
		{"Invalid - scheduled job not found by name", notFoundName, true, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			e := echo.New()
			reqPath := fmt.Sprintf("%s/%s", common.ApiScheduleJobByNameRoute, testCase.jobName)
			req, err := http.NewRequest(http.MethodGet, reqPath, http.NoBody)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			c := e.NewContext(req, recorder)
			c.SetParamNames(common.Name)
			c.SetParamValues(testCase.jobName)
			err = controller.ScheduleJobByName(c)
			require.NoError(t, err)

			// Assert
			if testCase.errorExpected {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			} else {
				var res responseDTO.ScheduleJobResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
				assert.Equal(t, testCase.jobName, res.ScheduleJob.Name, "Name not as expected")
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			}
		})
	}
}

func TestAllScheduleJobs(t *testing.T) {
	expectedTotalScheduleJobsCount := int64(0)
	dic := mockDic()
	var emptyLabels []string
	dbClientMock := &csMock.DBClient{}
	dbClientMock.On("ScheduleJobTotalCount", context.Background(), emptyLabels).Return(expectedTotalScheduleJobsCount, nil)
	dbClientMock.On("ScheduleJobTotalCount", context.Background(), testScheduleJobLabels).Return(expectedTotalScheduleJobsCount, nil)
	dbClientMock.On("AllScheduleJobs", context.Background(), emptyLabels, 0, 20).Return([]models.ScheduleJob{}, nil)
	dbClientMock.On("AllScheduleJobs", context.Background(), emptyLabels, 0, 1).Return([]models.ScheduleJob{}, nil)
	dbClientMock.On("AllScheduleJobs", context.Background(), testScheduleJobLabels, 0, 1).Return([]models.ScheduleJob{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) any {
			return dbClientMock
		},
	})
	controller := NewScheduleJobController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		labels             string
		offset             string
		limit              string
		errorExpected      bool
		expectedTotalCount int64
		expectedStatusCode int
	}{
		{"Valid - get scheduled jobs without offset and limit", "", "", "", false, expectedTotalScheduleJobsCount, http.StatusOK},
		{"Valid - get scheduled jobs with offset and limit", "", "0", "1", false, expectedTotalScheduleJobsCount, http.StatusOK},
		{"Valid - get scheduled jobs by labels", strings.Join(testScheduleJobLabels, ","), "0", "1", false, expectedTotalScheduleJobsCount, http.StatusOK},
		{"Invalid - invalid offset format", "", "aaa", "1", true, expectedTotalScheduleJobsCount, http.StatusBadRequest},
		{"Invalid - invalid limit format", "", "1", "aaa", true, expectedTotalScheduleJobsCount, http.StatusBadRequest},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			e := echo.New()
			req, err := http.NewRequest(http.MethodGet, common.ApiAllScheduleJobRoute, http.NoBody)
			query := req.URL.Query()
			if testCase.labels != "" {
				query.Add(common.Labels, testCase.labels)
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
			err = controller.AllScheduleJobs(c)
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
				var res responseDTO.MultiScheduleJobsResponse
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

func TestDeleteScheduleJobByName(t *testing.T) {
	job := dtos.ToScheduleJobModel(addScheduleJobRequestData().ScheduleJob)
	noName := ""
	notFoundName := "notFoundName"

	dic := mockDic()
	dbClientMock := &csMock.DBClient{}
	schedulerManagerMock := &csMock.SchedulerManager{}
	dbClientMock.On("DeleteScheduleJobByName", context.Background(), job.Name).Return(nil)
	schedulerManagerMock.On("DeleteScheduleJobByName", job.Name, testCorrelationID).Return(nil)
	schedulerManagerMock.On("DeleteScheduleJobByName", noName, testCorrelationID).Return(errors.NewCommonEdgeX(errors.KindContractInvalid, "scheduled job name is required", nil))
	schedulerManagerMock.On("DeleteScheduleJobByName", notFoundName, testCorrelationID).Return(errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "scheduled job doesn't exist in the scheduler manager", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) any {
			return dbClientMock
		},
		container.SchedulerManagerName: func(get di.Get) any {
			return schedulerManagerMock
		},
	})

	controller := NewScheduleJobController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		jobName            string
		expectedStatusCode int
	}{
		{"Valid - scheduled job by name", job.Name, http.StatusOK},
		{"Invalid - name parameter is empty", noName, http.StatusBadRequest},
		{"Invalid - scheduled job not found by name", notFoundName, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			e := echo.New()
			reqPath := fmt.Sprintf("%s/%s", common.ApiTriggerScheduleJobByNameRoute, testCase.jobName)
			req, err := http.NewRequest(http.MethodDelete, reqPath, http.NoBody)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			c := e.NewContext(req, recorder)
			c.SetParamNames(common.Name)
			c.SetParamValues(testCase.jobName)
			err = controller.DeleteScheduleJobByName(c)
			require.NoError(t, err)
			var res commonDTO.BaseResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			require.NoError(t, err)

			// Assert
			assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
			assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, testCase.expectedStatusCode, int(res.StatusCode), "Response status code not as expected")
			if testCase.expectedStatusCode == http.StatusOK {
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			} else {
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			}
		})
	}
}

func TestPatchScheduleJob(t *testing.T) {
	expectedRequestId := exampleUUID
	dic := mockDic()
	dbClientMock := &csMock.DBClient{}
	schedulerManagerMock := &csMock.SchedulerManager{}
	testReq := updateScheduleJobRequestData()
	model := models.ScheduleJob{
		Id:         *testReq.ScheduleJob.Id,
		Name:       *testReq.ScheduleJob.Name,
		Definition: dtos.ToScheduleDefModel(*testReq.ScheduleJob.Definition),
		Actions:    dtos.ToScheduleActionModels(testReq.ScheduleJob.Actions),
		Labels:     testReq.ScheduleJob.Labels,
		Properties: testReq.ScheduleJob.Properties,
	}

	valid := testReq
	dbClientMock.On("ScheduleJobById", context.Background(), *valid.ScheduleJob.Id).Return(model, nil)
	dbClientMock.On("UpdateScheduleJob", context.Background(), mock.MatchedBy(func(job models.ScheduleJob) bool { return job.Name == testScheduleJobName })).Return(nil)
	schedulerManagerMock.On("UpdateScheduleJob", mock.MatchedBy(func(job models.ScheduleJob) bool { return job.Name == testScheduleJobName }), testCorrelationID).Return(nil)
	validWithNoReqID := testReq
	validWithNoReqID.RequestId = ""
	validWithNoId := testReq
	validWithNoId.ScheduleJob.Id = nil
	dbClientMock.On("ScheduleJobByName", context.Background(), *validWithNoId.ScheduleJob.Name).Return(model, nil)
	validWithNoName := testReq
	validWithNoName.ScheduleJob.Name = nil

	invalidId := testReq
	invalidUUID := "invalidUUID"
	invalidId.ScheduleJob.Id = &invalidUUID

	emptyString := ""
	emptyId := testReq
	emptyId.ScheduleJob.Id = &emptyString
	emptyId.ScheduleJob.Name = nil
	emptyName := testReq
	emptyName.ScheduleJob.Id = nil
	emptyName.ScheduleJob.Name = &emptyString

	invalidNoIdAndName := testReq
	invalidNoIdAndName.ScheduleJob.Id = nil
	invalidNoIdAndName.ScheduleJob.Name = nil

	invalidNotFoundId := testReq
	invalidNotFoundId.ScheduleJob.Name = nil
	notFoundId := "12345678-1111-1234-5678-de9dac3fb9bc"
	invalidNotFoundId.ScheduleJob.Id = &notFoundId
	notFoundIdError := errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("%s doesn't exist in the database", notFoundId), nil)
	dbClientMock.On("ScheduleJobById", context.Background(), *invalidNotFoundId.ScheduleJob.Id).Return(model, notFoundIdError)

	invalidNotFoundName := testReq
	invalidNotFoundName.ScheduleJob.Id = nil
	notFoundName := "notFoundName"
	invalidNotFoundName.ScheduleJob.Name = &notFoundName
	notFoundNameError := errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("%s doesn't exist in the database", notFoundName), nil)
	dbClientMock.On("ScheduleJobByName", context.Background(), *invalidNotFoundName.ScheduleJob.Name).Return(model, notFoundNameError)

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) any {
			return dbClientMock
		},
		container.SchedulerManagerName: func(get di.Get) any {
			return schedulerManagerMock
		},
	})
	controller := NewScheduleJobController(dic)
	require.NotNil(t, controller)
	tests := []struct {
		name                 string
		request              []requests.UpdateScheduleJobRequest
		expectedStatusCode   int
		expectedResponseCode int
	}{
		{"Valid", []requests.UpdateScheduleJobRequest{valid}, http.StatusMultiStatus, http.StatusOK},
		{"Valid - no requestId", []requests.UpdateScheduleJobRequest{validWithNoReqID}, http.StatusMultiStatus, http.StatusOK},
		{"Valid - no id", []requests.UpdateScheduleJobRequest{validWithNoId}, http.StatusMultiStatus, http.StatusOK},
		{"Valid - no name", []requests.UpdateScheduleJobRequest{validWithNoName}, http.StatusMultiStatus, http.StatusOK},
		{"Invalid - invalid id", []requests.UpdateScheduleJobRequest{invalidId}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - empty id", []requests.UpdateScheduleJobRequest{emptyId}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - empty name", []requests.UpdateScheduleJobRequest{emptyName}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - no id and name", []requests.UpdateScheduleJobRequest{invalidNoIdAndName}, http.StatusBadRequest, http.StatusBadRequest},
		{"Invalid - not found id", []requests.UpdateScheduleJobRequest{invalidNotFoundId}, http.StatusMultiStatus, http.StatusNotFound},
		{"Invalid - not found name", []requests.UpdateScheduleJobRequest{invalidNotFoundName}, http.StatusMultiStatus, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			e := echo.New()
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req, err := http.NewRequest(http.MethodPatch, common.ApiScheduleJobRoute, reader)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			c := e.NewContext(req, recorder)
			err = controller.PatchScheduleJob(c)
			require.NoError(t, err)

			if testCase.expectedStatusCode == http.StatusMultiStatus {
				var res []commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, http.StatusMultiStatus, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, common.ApiVersion, res[0].ApiVersion, "API Version not as expected")
				if res[0].RequestId != "" {
					assert.Equal(t, expectedRequestId, res[0].RequestId, "RequestID not as expected")
				}
				assert.Equal(t, testCase.expectedResponseCode, res[0].StatusCode, "BaseResponse status code not as expected")
				if testCase.expectedResponseCode == http.StatusOK {
					assert.Empty(t, res[0].Message, "Message should be empty when it is successful")
				} else {
					assert.NotEmpty(t, res[0].Message, "Response message doesn't contain the error message")
				}
			} else {
				var res commonDTO.BaseResponse
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Assert
				assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
				assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
				assert.Equal(t, testCase.expectedResponseCode, res.StatusCode, "BaseResponse status code not as expected")
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			}

		})
	}

}

func TestTriggerScheduleJobByName(t *testing.T) {
	job := dtos.ToScheduleJobModel(addScheduleJobRequestData().ScheduleJob)
	emptyName := ""
	notFoundName := "notFoundName"

	dic := mockDic()
	dbClientMock := &csMock.DBClient{}
	schedulerManagerMock := &csMock.SchedulerManager{}
	schedulerManagerMock.On("TriggerScheduleJobByName", job.Name, testCorrelationID).Return(nil)
	schedulerManagerMock.On("TriggerScheduleJobByName", notFoundName, testCorrelationID).Return(errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "scheduled job doesn't exist in the scheduler manager", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) any {
			return dbClientMock
		},
		container.SchedulerManagerName: func(get di.Get) any {
			return schedulerManagerMock
		},
	})

	controller := NewScheduleJobController(dic)
	require.NotNil(t, controller)

	tests := []struct {
		name               string
		jobName            string
		errorExpected      bool
		expectedStatusCode int
	}{
		{"Valid - trigger scheduled job by name", job.Name, false, http.StatusAccepted},
		{"Invalid - name parameter is empty", emptyName, true, http.StatusBadRequest},
		{"Invalid - scheduled job not found by name", notFoundName, true, http.StatusNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			e := echo.New()
			reqPath := fmt.Sprintf("%s/%s", common.ApiTriggerScheduleJobByNameRoute, testCase.jobName)
			req, err := http.NewRequest(http.MethodPost, reqPath, http.NoBody)
			require.NoError(t, err)

			// Act
			recorder := httptest.NewRecorder()
			c := e.NewContext(req, recorder)
			c.SetParamNames(common.Name)
			c.SetParamValues(testCase.jobName)
			err = controller.TriggerScheduleJobByName(c)
			require.NoError(t, err)

			var res commonDTO.BaseResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			require.NoError(t, err)

			// Assert
			assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
			if testCase.expectedStatusCode == http.StatusAccepted {
				assert.Empty(t, res.Message, "Message should be empty when the request is accepted")
			} else {
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			}
		})
	}
}
