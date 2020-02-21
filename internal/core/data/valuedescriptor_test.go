package data

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/data/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/data/interfaces/mocks"
	dataMocks "github.com/edgexfoundry/edgex-go/internal/core/data/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	"github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/stretchr/testify/mock"
)

func TestValidateFormatString(t *testing.T) {
	err := validateFormatString(models.ValueDescriptor{Formatting: "%s"}, logger.NewMockClient())

	if err != nil {
		t.Errorf("Should match format specifier")
	}
}

func TestValidateFormatStringEmpty(t *testing.T) {
	err := validateFormatString(models.ValueDescriptor{Formatting: ""}, logger.NewMockClient())

	if err != nil {
		t.Errorf("Should match format specifier")
	}
}

func TestValidateFormatStringInvalid(t *testing.T) {
	err := validateFormatString(models.ValueDescriptor{Formatting: "error"}, logger.NewMockClient())

	if err == nil {
		t.Errorf("Expected error on invalid format string")
	}
}

func TestGetValueDescriptorByName(t *testing.T) {
	reset()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ValueDescriptorByName", mock.Anything).Return(models.ValueDescriptor{Id: testUUIDString}, nil)
	valueDescriptor, err := getValueDescriptorByName("valid", logger.NewMockClient(), dbClientMock)
	if err != nil {
		t.Errorf("Unexpected error getting value descriptor by name")
	}

	if valueDescriptor.Id != testUUIDString {
		t.Errorf("ID returned doesn't match db")
	}
}

func TestGetValueDescriptorByNameNotFound(t *testing.T) {
	reset()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ValueDescriptorByName", mock.Anything).Return(models.ValueDescriptor{}, db.ErrNotFound)
	_, err := getValueDescriptorByName("404", logger.NewMockClient(), dbClientMock)
	if err != nil {
		switch err.(type) {
		case errors.ErrDbNotFound:
			return
		default:
			t.Errorf("Unexpected error getting value descriptor by name missing in DB")
		}
	}

	if err == nil {
		t.Errorf("Expected error getting value descriptor by name missing in DB")
	}
}

func TestGetValueDescriptorByNameError(t *testing.T) {
	reset()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ValueDescriptorByName", mock.Anything).Return(models.ValueDescriptor{}, fmt.Errorf("some error"))
	_, err := getValueDescriptorByName("error", logger.NewMockClient(), dbClientMock)
	if err == nil {
		t.Errorf("Expected error getting value descriptor by name with some error")
	}
}

func TestGetValueDescriptorById(t *testing.T) {
	reset()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ValueDescriptorById", mock.Anything).Return(models.ValueDescriptor{Id: testUUIDString}, nil)
	valueDescriptor, err := getValueDescriptorById("valid", logger.NewMockClient(), dbClientMock)
	if err != nil {
		t.Errorf("Unexpected error getting value descriptor by ID")
	}

	if valueDescriptor.Id != testUUIDString {
		t.Errorf("ID returned doesn't match db")
	}
}

func TestGetValueDescriptorByIdNotFound(t *testing.T) {
	reset()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ValueDescriptorById", mock.Anything).Return(models.ValueDescriptor{}, db.ErrNotFound)
	_, err := getValueDescriptorById("404", logger.NewMockClient(), dbClientMock)
	if err != nil {
		switch err.(type) {
		case errors.ErrDbNotFound:
			return
		default:
			t.Errorf("Unexpected error getting value descriptor by ID missing in DB")
		}
	}

	if err == nil {
		t.Errorf("Expected error getting value descriptor by ID missing in DB")
	}
}

func TestGetValueDescriptorByIdError(t *testing.T) {
	reset()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ValueDescriptorById", mock.Anything).Return(models.ValueDescriptor{}, fmt.Errorf("some error"))
	_, err := getValueDescriptorById("error", logger.NewMockClient(), dbClientMock)
	if err == nil {
		t.Errorf("Expected error getting value descriptor by ID with some error")
	}
}

func TestGetValueDescriptorsByUomLabel(t *testing.T) {
	reset()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ValueDescriptorsByUomLabel", mock.Anything).Return([]models.ValueDescriptor{}, nil)
	_, err := getValueDescriptorsByUomLabel("valid", logger.NewMockClient(), dbClientMock)
	if err != nil {
		t.Errorf("Unexpected error getting value descriptor by UOM label")
	}
}

func TestGetValueDescriptorsByUomLabelNotFound(t *testing.T) {
	reset()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ValueDescriptorsByUomLabel", mock.Anything).Return([]models.ValueDescriptor{}, db.ErrNotFound)
	_, err := getValueDescriptorsByUomLabel("404", logger.NewMockClient(), dbClientMock)
	if err != nil {
		switch err.(type) {
		case errors.ErrDbNotFound:
			return
		default:
			t.Errorf("Unexpected error getting value descriptor by UOM label missing in DB")
		}
	}

	if err == nil {
		t.Errorf("Expected error getting value descriptor by UOM label missing in DB")
	}
}

func TestGetValueDescriptorsByUomLabelError(t *testing.T) {
	reset()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ValueDescriptorsByUomLabel", mock.Anything).Return([]models.ValueDescriptor{}, fmt.Errorf("some error"))
	_, err := getValueDescriptorsByUomLabel("error", logger.NewMockClient(), dbClientMock)
	if err == nil {
		t.Errorf("Expected error getting value descriptor by UOM label with some error")
	}
}

func TestGetValueDescriptorsByLabel(t *testing.T) {
	reset()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ValueDescriptorsByLabel", mock.MatchedBy(func(name string) bool {
		return name == testUUIDString
	})).Return([]models.ValueDescriptor{{Id: testUUIDString}}, nil)
	valueDescriptor, err := getValueDescriptorsByLabel(testUUIDString, logger.NewMockClient(), dbClientMock)
	if err != nil {
		t.Errorf("Unexpected error getting value descriptor by label")
	}

	if valueDescriptor[0].Id != testUUIDString {
		t.Errorf("ValueDescriptor received doesn't match expectation")
	}
}

func TestGetValueDescriptorsByLabelNotFound(t *testing.T) {
	reset()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ValueDescriptorsByLabel", mock.Anything).Return([]models.ValueDescriptor{}, db.ErrNotFound)
	_, err := getValueDescriptorsByLabel("404", logger.NewMockClient(), dbClientMock)
	if err != nil {
		switch err.(type) {
		case errors.ErrDbNotFound:
			return
		default:
			t.Errorf("Unexpected error getting value descriptor by label missing in DB")
		}
	}

	if err == nil {
		t.Errorf("Expected error getting value descriptor by label missing in DB")
	}
}

func TestGetValueDescriptorsByLabelError(t *testing.T) {
	reset()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ValueDescriptorsByLabel", mock.Anything).Return([]models.ValueDescriptor{}, fmt.Errorf("some error"))
	_, err := getValueDescriptorsByLabel("error", logger.NewMockClient(), dbClientMock)
	if err == nil {
		t.Errorf("Expected error getting value descriptor by label with some error")
	}
}

func TestGetValueDescriptorsByType(t *testing.T) {
	reset()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ValueDescriptorsByType", mock.Anything).Return([]models.ValueDescriptor{}, nil)
	_, err := getValueDescriptorsByType("valid", logger.NewMockClient(), dbClientMock)
	if err != nil {
		t.Errorf("Unexpected error getting value descriptor by type")
	}
}

func TestGetValueDescriptorsByTypeNotFound(t *testing.T) {
	reset()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ValueDescriptorsByType", mock.Anything).Return([]models.ValueDescriptor{}, db.ErrNotFound)
	_, err := getValueDescriptorsByType("404", logger.NewMockClient(), dbClientMock)
	if err != nil {
		switch err.(type) {
		case errors.ErrDbNotFound:
			return
		default:
			t.Errorf("Unexpected error getting value descriptor by type missing in DB")
		}
	}

	if err == nil {
		t.Errorf("Expected error getting value descriptor by type missing in DB")
	}
}

func TestGetValueDescriptorsByTypeError(t *testing.T) {
	reset()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ValueDescriptorsByType", mock.Anything).Return([]models.ValueDescriptor{}, fmt.Errorf("some error"))
	_, err := getValueDescriptorsByType("R", logger.NewMockClient(), dbClientMock)

	if err == nil {
		t.Errorf("Expected error getting value descriptor by type with some error")
	}
}

func TestGetValueDescriptorsByDeviceName(t *testing.T) {
	reset()
	_, err := getValueDescriptorsByDeviceName(context.Background(), testDeviceName, logger.NewMockClient(), nil, dataMocks.NewMockDeviceClient())
	if err != nil {
		t.Errorf("Unexpected error getting value descriptor by device name")
	}
}

func TestGetValueDescriptorsByDeviceNameNotFound(t *testing.T) {
	reset()
	_, err := getValueDescriptorsByDeviceName(context.Background(), "404", logger.NewMockClient(), nil, dataMocks.NewMockDeviceClient())
	if err != nil {
		switch err := err.(type) {
		case types.ErrServiceClient:
			if err.StatusCode != http.StatusNotFound {
				t.Errorf("Expected a 404 error")
			}
			return
		default:
			t.Errorf("Unexpected error getting value descriptor by device name missing in DB")
		}
	}

	if err == nil {
		t.Errorf("Expected error getting value descriptor by device name missing in DB")
	}
}

func TestGetValueDescriptorsByDeviceNameError(t *testing.T) {
	reset()
	_, err := getValueDescriptorsByDeviceName(context.Background(), "error", logger.NewMockClient(), nil, dataMocks.NewMockDeviceClient())
	if err == nil {
		t.Errorf("Expected error getting value descriptor by device name with some error")
	}
}

func TestGetValueDescriptorsByDeviceId(t *testing.T) {
	reset()
	_, err := getValueDescriptorsByDeviceId(context.Background(), "valid", logger.NewMockClient(), nil, dataMocks.NewMockDeviceClient())
	if err != nil {
		t.Errorf("Unexpected error getting value descriptor by device id")
	}
}

func TestGetValueDescriptorsByDeviceIdNotFound(t *testing.T) {
	reset()
	_, err := getValueDescriptorsByDeviceId(context.Background(), "404", logger.NewMockClient(), nil, dataMocks.NewMockDeviceClient())
	if err != nil {
		switch err := err.(type) {
		case types.ErrServiceClient:
			if err.StatusCode != http.StatusNotFound {
				t.Errorf("Expected a 404 error")
			}
			return
		default:
			t.Errorf("Unexpected error getting value descriptor by device id missing in DB")
		}
	}

	if err == nil {
		t.Errorf("Expected error getting value descriptor by device name missing in DB")
	}
}

func TestGetValueDescriptorsByDeviceIdError(t *testing.T) {
	reset()
	_, err := getValueDescriptorsByDeviceId(context.Background(), "error", logger.NewMockClient(), nil, dataMocks.NewMockDeviceClient())
	if err == nil {
		t.Errorf("Expected error getting value descriptor by device id with some error")
	}
}

func TestGetAllValueDescriptors(t *testing.T) {
	reset()
	vds := []models.ValueDescriptor{
		{Id: testUUIDString},
		{Id: testBsonString},
	}

	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ValueDescriptors").Return(vds, nil)
	_, err := getAllValueDescriptors(logger.NewMockClient(), dbClientMock)
	if err != nil {
		t.Errorf("Unexpected error getting all value descriptors")
	}
}

func TestGetAllValueDescriptorsError(t *testing.T) {
	reset()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ValueDescriptors").Return([]models.ValueDescriptor{}, fmt.Errorf("some error"))
	_, err := getAllValueDescriptors(logger.NewMockClient(), dbClientMock)
	if err == nil {
		t.Errorf("Expected error getting all value descriptors some error")
	}
}

func TestAddValueDescriptor(t *testing.T) {
	reset()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("AddValueDescriptor", mock.Anything).Return("", nil)
	_, err := addValueDescriptor(models.ValueDescriptor{Name: "valid"}, logger.NewMockClient(), dbClientMock)
	if err != nil {
		t.Errorf("Unexpected error adding value descriptor")
	}
}

func TestAddDuplicateValueDescriptor(t *testing.T) {
	reset()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("AddValueDescriptor", mock.Anything).Return("", db.ErrNotUnique)
	_, err := addValueDescriptor(models.ValueDescriptor{Name: "409"}, logger.NewMockClient(), dbClientMock)
	if err != nil {
		switch err.(type) {
		case errors.ErrDuplicateValueDescriptorName:
			return
		default:
			t.Errorf("Unexpected error adding value descriptor that already exists")
		}
	}

	if err == nil {
		t.Errorf("Expected error adding value descriptor that already exists")
	}
}

func TestAddValueDescriptorError(t *testing.T) {
	reset()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("AddValueDescriptor", mock.Anything).Return("", fmt.Errorf("some error"))
	_, err := addValueDescriptor(models.ValueDescriptor{}, logger.NewMockClient(), dbClientMock)
	if err == nil {
		t.Errorf("Expected error adding value descriptor some error")
	}
}

func TestDeleteValueDescriptor(t *testing.T) {
	reset()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("DeleteValueDescriptorById", mock.Anything).Return(nil)
	dbClientMock.On("ReadingsByValueDescriptor", mock.Anything, mock.Anything).Return([]models.Reading{}, nil)
	err := deleteValueDescriptor(models.ValueDescriptor{Name: "valid", Id: testBsonString}, logger.NewMockClient(), dbClientMock)
	if err != nil {
		t.Errorf("Unexpected error deleting value descriptor")
	}
}

func TestDeleteValueDescriptorInUse(t *testing.T) {
	reset()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ReadingsByValueDescriptor", mock.Anything, mock.Anything).Return([]models.Reading{{Id: testUUIDString}}, nil)
	err := deleteValueDescriptor(models.ValueDescriptor{Name: "409"}, logger.NewMockClient(), dbClientMock)
	if err != nil {
		switch err.(type) {
		case errors.ErrValueDescriptorInUse:
			return
		default:
			t.Errorf("Unexpected error deleting value descriptor in use")
		}
	}

	if err == nil {
		t.Errorf("Expected error deleting value descriptor in use")
	}
}

func TestDeleteValueDescriptorErrorReadingsLookup(t *testing.T) {
	reset()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ReadingsByValueDescriptor", mock.Anything, mock.Anything).Return([]models.Reading{}, fmt.Errorf("some error"))
	err := deleteValueDescriptor(models.ValueDescriptor{}, logger.NewMockClient(), dbClientMock)
	if err == nil {
		t.Errorf("Expected error deleting value descriptor some error looking up readings")
	}
}

func TestDeleteValueDescriptorError(t *testing.T) {
	reset()
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ReadingsByValueDescriptor", mock.Anything, mock.Anything).Return([]models.Reading{}, nil)
	dbClientMock.On("DeleteValueDescriptorById", mock.Anything).Return(fmt.Errorf("some error"))
	err := deleteValueDescriptor(models.ValueDescriptor{Name: "validErrorTest"}, logger.NewMockClient(), dbClientMock)
	if err == nil {
		t.Errorf("Expected error deleting value descriptor some error")
	}
}

type closingBuffer struct {
	*bytes.Buffer
}

func (cb *closingBuffer) Close() (err error) {
	return nil
}
