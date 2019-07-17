package addressable

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/addressable/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/stretchr/testify/mock"
)

var missingName = contract.Addressable{
	User:       "User1",
	Protocol:   "http",
	Id:         "address #1",
	HTTPMethod: "POST",
}

var duplicate = contract.Addressable{
	User:       "faker",
	Name:       "Not Good",
	Protocol:   "http",
	Id:         "duplicated",
	HTTPMethod: "POST",
}

func TestExecutor(t *testing.T) {
	success := SuccessfulDatabaseResult[0]

	tests := []struct {
		name             string
		mockDb           AddressWriter
		addr             contract.Addressable
		expectedResult   string
		expectedError    bool
		expectedErrorVal error
	}{
		{
			name:             "Successful database call",
			mockDb:           createAddMockAddressWriter(nil, success.Id),
			addr:             success,
			expectedResult:   success.Id,
			expectedError:    false,
			expectedErrorVal: nil,
		},
		{
			name:             "Unsuccessful database call",
			mockDb:           createAddMockAddressWriter(Error, ""),
			addr:             success,
			expectedResult:   "",
			expectedError:    true,
			expectedErrorVal: Error,
		},
		{
			name:             "Missing name field on new addressable",
			mockDb:           createAddMockAddressWriter(nil, missingName.Id),
			addr:             missingName,
			expectedResult:   "",
			expectedError:    true,
			expectedErrorVal: errors.NewErrEmptyAddressableName(),
		},
		{
			name:             "Duplicated addressable",
			mockDb:           createAddMockAddressWriter(Error, duplicate.Id),
			addr:             duplicate,
			expectedResult:   "",
			expectedError:    true,
			expectedErrorVal: errors.NewErrDuplicateName(fmt.Sprintf("duplicate name for addressable: %s", duplicate.Name)),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			op := NewAddExecutor(test.mockDb, test.addr)
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

			if test.expectedErrorVal != nil && err != nil {
				if test.expectedErrorVal.Error() != err.Error() {
					t.Errorf("Observed error doesn't match expected.\nExpected: %v\nActual: %v\n", test.expectedErrorVal.Error(), err.Error())
				}
			}
		})
	}
}

func createAddMockAddressWriter(err error, id string) AddressWriter {
	dbMock := mocks.AddressLoader{}
	dbMock.On("AddAddressable", duplicate).Return(id, db.ErrNotUnique)
	dbMock.On("AddAddressable", mock.Anything).Return(id, err)
	return &dbMock
}
