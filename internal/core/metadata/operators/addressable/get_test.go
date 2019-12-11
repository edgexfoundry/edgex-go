package addressable

import (
	"reflect"
	"testing"

	metadataErrors "github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/addressable/mocks"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/pkg/errors"
)

var Id = "83cb038b-5a94-4707-985d-13effec62de2"
var AddressName = "TestAddress"
var PublisherName = "TestPublisher"
var Topic = "TestTopic"
var Error = errors.New("test error")
var Port = 8080
var SuccessfulDatabaseResult = []contract.Addressable{
	{
		User:       "User1",
		Name:       "Ein",
		Protocol:   "http",
		Id:         "address #1",
		HTTPMethod: "POST",
		Address:    "localhost",
		Port:       1337,
		Path:       "/tests",
		Publisher:  "Brandon!",
		Password:   "hunter2",
		Topic:      "Hot",
	},
	{
		User:       "User2",
		Name:       "Zwei",
		Protocol:   "http",
		Id:         "address #2",
		HTTPMethod: "GET",
		Address:    "localhost",
		Port:       1337,
		Path:       "/tests",
		Publisher:  "Brandon!",
		Password:   "hunter2",
		Topic:      "Hot",
	},
}

func TestAddressExecutor(t *testing.T) {
	tests := []struct {
		name           string
		mockDb         AddressLoader
		expectedResult []contract.Addressable
		expectedError  bool
	}{
		{
			name:           "Successful database call",
			mockDb:         createMockAddressLoaderStringArg("GetAddressablesByAddress", nil, SuccessfulDatabaseResult, AddressName),
			expectedResult: SuccessfulDatabaseResult,
			expectedError:  false,
		},
		{
			name:           "Error database result",
			mockDb:         createMockAddressLoaderStringArg("GetAddressablesByAddress", Error, nil, AddressName),
			expectedResult: nil,
			expectedError:  true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewAddressExecutor(test.mockDb, AddressName)
			actual, err := op.Execute()
			if test.expectedError && err == nil {
				t.Error("Expected an error")
				return
			}

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			if !reflect.DeepEqual(test.expectedResult, actual) {
				t.Errorf("Expected result does not match the observed. \nExpected : %v \n Observed: %v", test.expectedResult, actual)
				return
			}
		})
	}
}

func TestAllAddressablesExecutor(t *testing.T) {
	tests := []struct {
		name             string
		mockDb           AddressLoader
		cfg              bootstrapConfig.ServiceInfo
		expectedResult   []contract.Addressable
		expectedError    bool
		expectedErrorVal error
	}{
		{
			name:           "Successful database call",
			mockDb:         createMockAddressLoader(nil, SuccessfulDatabaseResult),
			cfg:            bootstrapConfig.ServiceInfo{MaxResultCount: 5},
			expectedResult: SuccessfulDatabaseResult,
			expectedError:  false,
		},
		{
			name:           "Unsuccessful database call",
			mockDb:         createMockAddressLoader(Error, nil),
			cfg:            bootstrapConfig.ServiceInfo{MaxResultCount: 5},
			expectedResult: nil,
			expectedError:  true,
		},
		{
			name:             "MaxResultCount exceeded",
			mockDb:           createMockAddressLoader(nil, SuccessfulDatabaseResult),
			cfg:              bootstrapConfig.ServiceInfo{MaxResultCount: 1},
			expectedResult:   nil,
			expectedError:    true,
			expectedErrorVal: metadataErrors.NewErrLimitExceeded(1),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewAddressableLoadAll(test.cfg, test.mockDb, logger.NewMockClient())
			actual, err := op.Execute()
			if test.expectedError && err == nil {
				t.Error("Expected an error")
				return
			}
			if !test.expectedError && err != nil {
				t.Error("Unexpected error")
				return
			}
			if !reflect.DeepEqual(actual, test.expectedResult) {
				t.Errorf("Observed result doesn't match expected.\nExpected: %v\nActual: %v\n", test.expectedResult, actual)
			}
			if test.expectedErrorVal != nil && err != nil {
				if test.expectedErrorVal.Error() != err.Error() {
					t.Errorf("Observed error doesn't match expected.\nExpected: %v\nActual: %v\n", test.expectedErrorVal.Error(), err.Error())
				}
			}
		})
	}
}

func TestNameExecutor(t *testing.T) {
	tests := []struct {
		name           string
		mockDb         AddressLoader
		expectedResult contract.Addressable
		expectedError  bool
	}{
		{
			name:           "Successful database call",
			mockDb:         createMockAddressLoaderStringArg("GetAddressableByName", nil, SuccessfulDatabaseResult[0], AddressName),
			expectedResult: SuccessfulDatabaseResult[0],
			expectedError:  false,
		},
		{
			name:           "Unsuccessful database call",
			mockDb:         createMockAddressLoaderStringArg("GetAddressableByName", Error, contract.Addressable{}, AddressName),
			expectedResult: contract.Addressable{},
			expectedError:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewNameExecutor(test.mockDb, AddressName)
			actual, err := op.Execute()
			if test.expectedError && err == nil {
				t.Error("Expected an error")
				return
			}

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			if !reflect.DeepEqual(test.expectedResult, actual) {
				t.Errorf("Expected result does not match the observed.\nExpected: %v\nObserved: %v\n", test.expectedResult, actual)
				return
			}
		})
	}
}

func TestIdExecutor(t *testing.T) {
	tests := []struct {
		name           string
		mockDb         AddressLoader
		expectedResult contract.Addressable
		expectedError  bool
	}{
		{
			name:           "Successful database call",
			mockDb:         createMockAddressLoaderStringArg("GetAddressableById", nil, SuccessfulDatabaseResult[0], Id),
			expectedResult: SuccessfulDatabaseResult[0],
			expectedError:  false,
		},
		{
			name:           "Unsuccessful database call",
			mockDb:         createMockAddressLoaderStringArg("GetAddressableById", Error, contract.Addressable{}, Id),
			expectedResult: contract.Addressable{},
			expectedError:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewIdExecutor(test.mockDb, Id)
			actual, err := op.Execute()
			if test.expectedError && err == nil {
				t.Error("Expected an error")
				return
			}

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			if !reflect.DeepEqual(test.expectedResult, actual) {
				t.Errorf("Expected result does not match the observed.\nExpected: %v\nObserved: %v\n", test.expectedResult, actual)
				return
			}
		})
	}
}

func TestPublisherExecutor(t *testing.T) {
	tests := []struct {
		name           string
		mockDb         AddressLoader
		expectedResult []contract.Addressable
		expectedError  bool
	}{
		{
			name:           "Successful database call",
			mockDb:         createMockAddressLoaderStringArg("GetAddressablesByPublisher", nil, SuccessfulDatabaseResult, PublisherName),
			expectedResult: SuccessfulDatabaseResult,
			expectedError:  false,
		},
		{
			name:           "Error database result",
			mockDb:         createMockAddressLoaderStringArg("GetAddressablesByPublisher", Error, nil, PublisherName),
			expectedResult: nil,
			expectedError:  true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewPublisherExecutor(test.mockDb, PublisherName)
			actual, err := op.Execute()
			if test.expectedError && err == nil {
				t.Error("Expected an error")
				return
			}

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			if !reflect.DeepEqual(test.expectedResult, actual) {
				t.Errorf("Expected result does not match the observed. \nExpected : %v \n Observed: %v", test.expectedResult, actual)
				return
			}
		})
	}
}

func TestPortExecutor(t *testing.T) {
	tests := []struct {
		name           string
		mockDb         AddressLoader
		expectedResult []contract.Addressable
		expectedError  bool
	}{
		{
			name:           "Successful database call",
			mockDb:         createMockAddressLoaderIntArg("GetAddressablesByPort", nil, SuccessfulDatabaseResult, Port),
			expectedResult: SuccessfulDatabaseResult,
			expectedError:  false,
		},
		{
			name:           "Error database result",
			mockDb:         createMockAddressLoaderIntArg("GetAddressablesByPort", Error, nil, Port),
			expectedResult: nil,
			expectedError:  true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewPortExecutor(test.mockDb, Port)
			actual, err := op.Execute()
			if test.expectedError && err == nil {
				t.Error("Expected an error")
				return
			}

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			if !reflect.DeepEqual(test.expectedResult, actual) {
				t.Errorf("Expected result does not match the observed. \nExpected : %v \n Observed: %v", test.expectedResult, actual)
				return
			}
		})
	}
}

func TestTopicExecutor(t *testing.T) {
	tests := []struct {
		name           string
		mockDb         AddressLoader
		expectedResult []contract.Addressable
		expectedError  bool
	}{
		{
			name:           "Successful database call",
			mockDb:         createMockAddressLoaderStringArg("GetAddressablesByTopic", nil, SuccessfulDatabaseResult, Topic),
			expectedResult: SuccessfulDatabaseResult,
			expectedError:  false,
		},
		{
			name:           "Error database result",
			mockDb:         createMockAddressLoaderStringArg("GetAddressablesByTopic", Error, nil, Topic),
			expectedResult: nil,
			expectedError:  true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewTopicExecutor(test.mockDb, Topic)
			actual, err := op.Execute()
			if test.expectedError && err == nil {
				t.Error("Expected an error")
				return
			}

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			if !reflect.DeepEqual(test.expectedResult, actual) {
				t.Errorf("Expected result does not match the observed. \nExpected : %v \n Observed: %v", test.expectedResult, actual)
				return
			}
		})
	}
}

func createMockAddressLoaderStringArg(methodName string, err error, ret interface{}, arg string) AddressLoader {
	dbMock := mocks.AddressLoader{}
	dbMock.On(methodName, arg).Return(ret, err)
	return &dbMock
}

func createMockAddressLoaderIntArg(methodName string, err error, ret interface{}, arg int) AddressLoader {
	dbMock := mocks.AddressLoader{}
	dbMock.On(methodName, arg).Return(ret, err)
	return &dbMock
}

func createMockAddressLoader(err error, ret []contract.Addressable) AddressLoader {
	dbMock := mocks.AddressLoader{}
	dbMock.On("GetAddressables").Return(ret, err)
	return &dbMock
}
