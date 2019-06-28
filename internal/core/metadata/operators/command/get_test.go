package command

import (
	"reflect"
	"testing"

	mock "github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/command/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

func TestLoadCommands(t *testing.T) {
	tests := []struct {
		name           string
		commandDBMock  CommandLoader
		expectedResult []models.Command
		expectError    bool
	}{
		{"GetCommandsByDeviceId", createCommandLoaderMock(), []models.Command{testCommand}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opp := NewDeviceIdExecutor(tt.commandDBMock, newDeviceId)
			commands, err := opp.Execute()
			if tt.expectError && err == nil {
				t.Error("Expected an error")
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			if !reflect.DeepEqual(tt.expectedResult, commands) {
				t.Errorf("Expected result does not match the observed. \nExpected : %v \n Observed: %v", tt.expectError, tt.expectedResult)
				return
			}

		})
	}
}

func createCommandLoaderMock() CommandLoader {
	commands := []models.Command{testCommand}
	dbCommandMoch := &mock.CommandLoader{}
	dbCommandMoch.On("GetCommandsByDeviceId", newDeviceId).Return(commands, nil)
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
