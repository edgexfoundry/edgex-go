package agent

import (
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	requests "github.com/edgexfoundry/go-mod-core-contracts/requests/configuration"
	responses "github.com/edgexfoundry/go-mod-core-contracts/responses/configuration"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetExecutorWithNoServices(t *testing.T) {

	loggingClient := logger.NewMockClient()
	executor := NewStubSet(stubCallSet{})
	sut := NewSetExecutor(executor.SetExecutor, loggingClient, &ConfigurationStruct{})
	sc := requests.SetConfigRequest{}
	actual := sut.Set([]string{}, sc)

	expectedResult := struct {
		Configuration map[string]responses.SetConfigResponse "json:\"configuration\""
	}(struct {
		Configuration map[string]responses.SetConfigResponse "json:\"configuration\""
	}{
		Configuration: map[string]responses.SetConfigResponse{}})

	assert.Equal(t, expectedResult, actual)
	assert.Equal(t, executor.Called, 0)
}

func TestSetExecutorWithServices(t *testing.T) {

	const (
		service1Name = "service1Name"
	)

	service1ExpectedResult := struct {
		Configuration map[string]responses.SetConfigResponse "json:\"configuration\""
	}{Configuration: map[string]responses.SetConfigResponse{"service1Name": {Success: false, Description: ""}}}

	loggingClient := logger.NewMockClient()
	sc := requests.SetConfigRequest{Key: "Writable.LogLevel", Value: "INFO"}

	tests := []struct {
		name           string
		services       []string
		expectedResult struct {
			Configuration map[string]responses.SetConfigResponse `json:"configuration"`
		}
		executorCalls stubCallSet
	}{
		{
			"one service is the target of the set operation",
			[]string{service1Name},
			service1ExpectedResult,
			stubCallSet{[]string{service1Name}, responses.SetConfigResponse{}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executor := NewStubSet(test.executorCalls)
			sut := NewSetExecutor(executor.SetExecutor, loggingClient, &ConfigurationStruct{})
			actualResult := sut.Set(test.services, sc)
			assert.Equal(t, test.expectedResult, actualResult)
		})
	}
}
