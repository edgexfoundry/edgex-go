package command

import (
	"reflect"
	"testing"

	mock "github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/command/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
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
		{"GetCommandsByDeviceId", createCommandByDeviceLoaderMock(), []models.Command{testCommand}, false},
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
		cfg            config.ServiceInfo
		commandDBMock  CommandLoader
		expectedResult []models.Command
		expectError    bool
	}{
		{"GetAllOK", config.ServiceInfo{MaxResultCount: 1}, createAllCommandsLoaderMock(), []models.Command{testCommand}, false},
		{"GetAllFailMaxExceeded", config.ServiceInfo{}, createAllCommandsLoaderMock(), []models.Command{testCommand}, true},
		{"GetAllFailUnexpected", config.ServiceInfo{}, createAllCommandsLoaderMockFail(), nil, true},
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

func createCommandByDeviceLoaderMock() CommandLoader {
	commands := []models.Command{testCommand}
	dbCommandMoch := &mock.CommandLoader{}
	dbCommandMoch.On("GetCommandsByDeviceId", newDeviceId).Return(commands, nil)
	return dbCommandMoch
}

func createAllCommandsLoaderMock() CommandLoader {
	commands := []models.Command{testCommand}
	dbCommandMoch := &mock.CommandLoader{}
	dbCommandMoch.On("GetAllCommands").Return(commands, nil)
	return dbCommandMoch
}

func createAllCommandsLoaderMockFail() CommandLoader {
	dbCommandMoch := &mock.CommandLoader{}
	dbCommandMoch.On("GetAllCommands").Return(nil, errors.New("unexpected error"))
	return dbCommandMoch
}

var newDeviceId = "b3445cc6-87df-48f4-b8b0-587dc8a4e1c2"
var testCommand = models.Command{Timestamps: testTimestamps, Name: "test command name", Get: models.Get{Action: testAction},
	Put: models.Put{Action: testAction, ParameterNames: testExpectedvalues}}

var testExpectedvalues = []string{"temperature", "humidity"}
var testAction = models.Action{Path: "test/path", Responses: []models.Response{{Code: "200", Description: "ok", ExpectedValues: testExpectedvalues}}, URL: ""}

//
var testTimestamps = models.Timestamps{Created: 123, Modified: 123, Origin: 123}
var testDescribedObject = models.DescribedObject{Timestamps: testTimestamps, Description: "This is a description"}
