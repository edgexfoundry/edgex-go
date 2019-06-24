package addressable

import (
	"reflect"
	"testing"

	"github.com/pkg/errors"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/addressable/mocks"
)

var AddressName = "TestAddress"
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

func TestAddressLoader_Execute(t *testing.T) {
	tests := []struct {
		name           string
		mockDb         AddressLoader
		expectedResult []contract.Addressable
		expectedError  bool
	}{
		{
			name:           "Successful database call",
			mockDb:         createMockAddressLoader(),
			expectedResult: SuccessfulDatabaseResult,
			expectedError:  false,
		},
		{
			name:           "Error database result",
			mockDb:         createErrorMockAddressLoader(),
			expectedResult: nil,
			expectedError:  true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewAddressLoader(test.mockDb, AddressName)
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

func createErrorMockAddressLoader() AddressLoader {
	dbMock := &mocks.AddressablesByAddressLoader{}
	dbMock.On("GetAddressablesByAddress", AddressName).Return(nil, errors.New("test error"))
	return dbMock
}

func createMockAddressLoader() AddressLoader {
	dbMock := &mocks.AddressablesByAddressLoader{}
	dbMock.On("GetAddressablesByAddress", AddressName).Return(SuccessfulDatabaseResult, nil)
	return dbMock
}
