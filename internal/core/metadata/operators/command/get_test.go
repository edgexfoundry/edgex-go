package command

import (
	"reflect"
	"testing"

	mock "github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/command/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"

	"github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/pkg/errors"
)

func TestLoadCommandsByDeviceId(t *testing.T) {
	tests := []struct {
		name           string
		commandDBMock  CommandLoader
		expectedResult []models.Command
		expectError    bool
	}{
		{
			"GetCommandsByDeviceIdOK",
			createCommandLoaderMock("GetCommandsByDeviceId", []models.Command{testCommand}, nil, newDeviceId),
			[]models.Command{testCommand},
			false,
		},
		{
			"GetCommandsByDeviceIdFailedNotFound",
			createCommandLoaderMock("GetCommandsByDeviceId", nil, db.ErrNotFound, newDeviceId),
			nil,
			true,
		},
		{
			"GetCommandsByDeviceIdFailedUnexpected",
			createCommandLoaderMock("GetCommandsByDeviceId", nil, errors.New("unexpected error"), newDeviceId),
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opp := NewDeviceIdExecutor(tt.commandDBMock, newDeviceId)
			result, err := opp.Execute()
			if tt.expectError && err == nil {
				t.Error("Expected an error")
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			if !reflect.DeepEqual(tt.expectedResult, result) {
				t.Errorf("Expected result does not match the observed. \nExpected : %v \n Observed: %v", tt.expectError, result)
				return
			}
		})
	}
}

func TestLoadAllCommands(t *testing.T) {
	tests := []struct {
		name           string
		cfg            bootstrapConfig.ServiceInfo
		commandDBMock  CommandLoader
		expectedResult []models.Command
		expectError    bool
	}{
		{
			"GetAllOK",
			bootstrapConfig.ServiceInfo{MaxResultCount: 1},
			createCommandLoaderMock("GetAllCommands", []models.Command{testCommand}, nil, ""),
			[]models.Command{testCommand},
			false,
		},
		{
			"GetAllFailMaxExceeded",
			bootstrapConfig.ServiceInfo{},
			createCommandLoaderMock("GetAllCommands", []models.Command{testCommand}, nil, ""),
			[]models.Command{},
			true,
		},
		{
			"GetAllFailUnexpected",
			bootstrapConfig.ServiceInfo{},
			createCommandLoaderMock("GetAllCommands", nil, errors.New("unexpected error"), ""),
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opp := NewCommandLoadAll(tt.cfg, tt.commandDBMock)
			result, err := opp.Execute()
			if tt.expectError && err == nil {
				t.Error("Expected an error")
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}
			if !reflect.DeepEqual(tt.expectedResult, result) {
				t.Errorf("Expected result does not match the observed. \nExpected : %v \n Observed: %v", tt.expectedResult, result)
				return
			}
		})
	}
}

func TestLoadCommandsById(t *testing.T) {
	tests := []struct {
		name           string
		commandDBMock  CommandLoader
		expectedResult models.Command
		expectError    bool
	}{
		{
			"GetCommandByIdOK",
			createCommandLoaderMock("GetCommandById", testCommand, nil, commandId),
			testCommand,
			false,
		},
		{
			"GetCommandByIdFailNotFound",
			createCommandLoaderMock("GetCommandById", models.Command{}, db.ErrNotFound, commandId),
			models.Command{},
			true,
		},
		{
			"GetCommandByIdFailUnexpected",
			createCommandLoaderMock("GetCommandById", models.Command{}, errors.New("unexpected error"), commandId),
			models.Command{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opp := NewCommandById(tt.commandDBMock, commandId)
			result, err := opp.Execute()
			if tt.expectError && err == nil {
				t.Error("Expected an error")
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			if !reflect.DeepEqual(tt.expectedResult, result) {
				t.Errorf("Expected result does not match the observed. \nExpected : %v \n Observed: %v", tt.expectError, result)
				return
			}

		})
	}
}

func TestLoadCommandsByName(t *testing.T) {
	tests := []struct {
		name           string
		commandDBMock  CommandLoader
		expectedResult []models.Command
		expectError    bool
	}{
		{
			"GetByNameOK",
			createCommandLoaderMock("GetCommandsByName", []models.Command{testCommand}, nil, commandName),
			[]models.Command{testCommand},
			false,
		},
		{
			"GetByNameFailUnexpected",
			createCommandLoaderMock("GetCommandsByName", nil, errors.New("unexpected error"), commandName),
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opp := NewCommandsByName(tt.commandDBMock, commandName)
			result, err := opp.Execute()
			if tt.expectError && err == nil {
				t.Error("Expected an error")
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}
			if !reflect.DeepEqual(tt.expectedResult, result) {
				t.Errorf("Expected result does not match the observed. \nExpected : %v \n Observed: %v", tt.expectedResult, result)
				return
			}
		})
	}
}

func createCommandLoaderMock(methodName string, ret interface{}, err error, arg string) CommandLoader {
	dbCommandMock := &mock.CommandLoader{}
	if arg != "" {
		dbCommandMock.On(methodName, arg).Return(ret, err)
	} else {
		dbCommandMock.On(methodName).Return(ret, err)
	}
	return dbCommandMock
}

var commandId = "f97b5f0a-ec32-4e96-bd36-02210af16f8c"
var commandName = "test command name"
var newDeviceId = "b3445cc6-87df-48f4-b8b0-587dc8a4e1c2"
var testCommand = models.Command{Timestamps: testTimestamps, Name: commandName, Get: models.Get{Action: testAction},
	Put: models.Put{Action: testAction, ParameterNames: testExpectedvalues}}

var testExpectedvalues = []string{"temperature", "humidity"}
var testAction = models.Action{Path: "test/path", Responses: []models.Response{{Code: "200", Description: "ok", ExpectedValues: testExpectedvalues}}, URL: ""}

var testTimestamps = models.Timestamps{Created: 123, Modified: 123, Origin: 123}
var testDescribedObject = models.DescribedObject{Timestamps: testTimestamps, Description: "This is a description"}
