package addressable

import (
	"reflect"
	"testing"

	"github.com/pkg/errors"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/addressable/mocks"
)

var AddressName = "TestAddress"
var PublisherName = "TestPublisher"
var Error = errors.New("test error")
var Port = 8080
var SuccessfulDatabaseResult = []contract.Addressable{
	{
		User:       "User1",
		Protocol:   "http",
		Id:         "address #1",
		HTTPMethod: "POST",
	},
	{
		User:       "User2",
		Protocol:   "http",
		Id:         "address #2",
		HTTPMethod: "GET",
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
			mockDb:         createMockAddressLoaderName("GetAddressablesByAddress", nil, SuccessfulDatabaseResult, AddressName),
			expectedResult: SuccessfulDatabaseResult,
			expectedError:  false,
		},
		{
			name:           "Error database result",
			mockDb:         createMockAddressLoaderName("GetAddressablesByAddress", Error, nil, AddressName),
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

func TestPublisherExecutor(t *testing.T) {
	tests := []struct {
		name           string
		mockDb         AddressLoader
		expectedResult []contract.Addressable
		expectedError  bool
	}{
		{
			name:           "Successful database call",
			mockDb:         createMockAddressLoaderName("GetAddressablesByPublisher", nil, SuccessfulDatabaseResult, PublisherName),
			expectedResult: SuccessfulDatabaseResult,
			expectedError:  false,
		},
		{
			name:           "Error database result",
			mockDb:         createMockAddressLoaderName("GetAddressablesByPublisher", Error, nil, PublisherName),
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
			mockDb:         createMockAddressLoaderPort("GetAddressablesByPort", nil, SuccessfulDatabaseResult, Port),
			expectedResult: SuccessfulDatabaseResult,
			expectedError:  false,
		},
		{
			name:           "Error database result",
			mockDb:         createMockAddressLoaderPort("GetAddressablesByPort", Error, nil, Port),
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

func createMockAddressLoaderName(methodName string, err error, ret interface{}, arg string) AddressLoader {
	dbMock := &mocks.AddressLoader{}
	dbMock.On(methodName, arg).Return(ret, err)
	return dbMock
}

func createMockAddressLoaderPort(methodName string, err error, ret interface{}, arg int) AddressLoader {
	dbMock := &mocks.AddressLoader{}
	dbMock.On(methodName, arg).Return(ret, err)
	return dbMock
}
