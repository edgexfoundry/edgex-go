package data

import (
	"bytes"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"net/http"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/data/errors"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

func TestValidateFormatString(t *testing.T) {
	err := validateFormatString(models.ValueDescriptor{Formatting: "%s"})

	if err != nil {
		t.Errorf("Should match format specifier")
	}
}

func TestValidateFormatStringEmpty(t *testing.T) {
	err := validateFormatString(models.ValueDescriptor{Formatting: ""})

	if err != nil {
		t.Errorf("Should match format specifier")
	}
}

func TestValidateFormatStringInvalid(t *testing.T) {
	err := validateFormatString(models.ValueDescriptor{Formatting: "error"})

	if err == nil {
		t.Errorf("Expected error on invalid format string")
	}
}

func TestGetValueDescriptorByName(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := getValueDescriptorByName("valid")

	if err != nil {
		t.Errorf("Unexpected error getting value descriptor by name")
	}
}

func TestGetValueDescriptorByNameNotFound(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := getValueDescriptorByName("404")

	if err != nil {
		switch err.(type) {
		case *errors.ErrDbNotFound:
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
	dbClient = newMockDb()

	_, err := getValueDescriptorByName("error")

	if err == nil {
		t.Errorf("Expected error getting value descriptor by name with some error")
	}
}

func TestGetValueDescriptorById(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := getValueDescriptorById("valid")

	if err != nil {
		t.Errorf("Unexpected error getting value descriptor by ID")
	}
}

func TestGetValueDescriptorByIdNotFound(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := getValueDescriptorById("404")

	if err != nil {
		switch err.(type) {
		case *errors.ErrDbNotFound:
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
	dbClient = newMockDb()

	_, err := getValueDescriptorById("error")

	if err == nil {
		t.Errorf("Expected error getting value descriptor by ID with some error")
	}
}

func TestGetValueDescriptorsByUomLabel(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := getValueDescriptorsByUomLabel("valid")

	if err != nil {
		t.Errorf("Unexpected error getting value descriptor by UOM label")
	}
}

func TestGetValueDescriptorsByUomLabelNotFound(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := getValueDescriptorsByUomLabel("404")

	if err != nil {
		switch err.(type) {
		case *errors.ErrDbNotFound:
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
	dbClient = newMockDb()

	_, err := getValueDescriptorsByUomLabel("error")

	if err == nil {
		t.Errorf("Expected error getting value descriptor by UOM label with some error")
	}
}

func TestGetValueDescriptorsByLabel(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := getValueDescriptorsByLabel("valid")

	if err != nil {
		t.Errorf("Unexpected error getting value descriptor by label")
	}
}

func TestGetValueDescriptorsByLabelNotFound(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := getValueDescriptorsByLabel("404")

	if err != nil {
		switch err.(type) {
		case *errors.ErrDbNotFound:
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
	dbClient = newMockDb()

	_, err := getValueDescriptorsByLabel("error")

	if err == nil {
		t.Errorf("Expected error getting value descriptor by label with some error")
	}
}

func TestGetValueDescriptorsByType(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := getValueDescriptorsByType("valid")

	if err != nil {
		t.Errorf("Unexpected error getting value descriptor by type")
	}
}

func TestGetValueDescriptorsByTypeNotFound(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := getValueDescriptorsByType("404")

	if err != nil {
		switch err.(type) {
		case *errors.ErrDbNotFound:
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
	dbClient = newMockDb()

	_, err := getValueDescriptorsByType("R")

	if err == nil {
		t.Errorf("Expected error getting value descriptor by type with some error")
	}
}

func TestGetValueDescriptorsByDeviceName(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := getValueDescriptorsByDeviceName(testDeviceName)

	if err != nil {
		t.Errorf("Unexpected error getting value descriptor by device name")
	}
}

func TestGetValueDescriptorsByDeviceNameNotFound(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := getValueDescriptorsByDeviceName("404")

	if err != nil {
		switch err := err.(type) {
		case *types.ErrServiceClient:
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
	dbClient = newMockDb()

	_, err := getValueDescriptorsByDeviceName("error")

	if err == nil {
		t.Errorf("Expected error getting value descriptor by device name with some error")
	}
}

func TestGetValueDescriptorsByDeviceId(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := getValueDescriptorsByDeviceId("valid")

	if err != nil {
		t.Errorf("Unexpected error getting value descriptor by device id")
	}
}

func TestGetValueDescriptorsByDeviceIdNotFound(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := getValueDescriptorsByDeviceId("404")

	if err != nil {
		switch err := err.(type) {
		case *types.ErrServiceClient:
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
	dbClient = newMockDb()

	_, err := getValueDescriptorsByDeviceId("error")

	if err == nil {
		t.Errorf("Expected error getting value descriptor by device id with some error")
	}
}

func TestGetAllValueDescriptors(t *testing.T) {
	reset()
	// this test uses the memdb on purpose

	_, err := getAllValueDescriptors()

	if err != nil {
		t.Errorf("Unexpected error getting all value descriptors")
	}
}

func TestGetAllValueDescriptorsError(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := getAllValueDescriptors()

	if err == nil {
		t.Errorf("Expected error getting all value descriptors some error")
	}
}

func TestAddValueDescriptor(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := addValueDescriptor(models.ValueDescriptor{Name: "valid"})

	if err != nil {
		t.Errorf("Unexpected error adding value descriptor")
	}
}

func TestAddValueDescriptorInUse(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := addValueDescriptor(models.ValueDescriptor{Name: "409"})

	if err != nil {
		switch err.(type) {
		case *errors.ErrValueDescriptorInUse:
			return
		default:
			t.Errorf("Unexpected error getting value descriptor by UOM label missing in DB")
		}
	}

	if err == nil {
		t.Errorf("Expected error adding value descriptor that already exists")
	}
}

func TestAddValueDescriptorError(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := addValueDescriptor(models.ValueDescriptor{})

	if err == nil {
		t.Errorf("Expected error adding value descriptor some error")
	}
}

func TestDeleteValueDescriptor(t *testing.T) {
	reset()
	dbClient = newMockDb()

	err := deleteValueDescriptor(models.ValueDescriptor{Name: "valid", Id: testBsonString})

	if err != nil {
		t.Errorf("Unexpected error deleting value descriptor")
	}
}

func TestDeleteValueDescriptorInUse(t *testing.T) {
	reset()
	dbClient = newMockDb()

	err := deleteValueDescriptor(models.ValueDescriptor{Name: "409"})

	if err != nil {
		switch err.(type) {
		case *errors.ErrValueDescriptorInUse:
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
	dbClient = newMockDb()

	err := deleteValueDescriptor(models.ValueDescriptor{})

	if err == nil {
		t.Errorf("Expected error deleting value descriptor some error looking up readings")
	}
}

func TestDeleteValueDescriptorError(t *testing.T) {
	reset()
	dbClient = newMockDb()

	err := deleteValueDescriptor(models.ValueDescriptor{Name: "validErrorTest"})

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

func TestDecodeValueDescriptorBadFormatting(t *testing.T) {
	reset()
	dbClient = newMockDb()

	cb := &closingBuffer{bytes.NewBufferString(`{
		"name": "co2",
		"min": "12",
		"max": "15",
		"type": "F",
		"uomLabel": "degreecel",
		"defaultValue": "0",
		"formatting": "%",
		"labels": [
			"NHCO2",
		"hvac"
	]
	}`)}
	_, err := decodeValueDescriptor(cb)

	switch err.(type) {
	case *errors.ErrValueDescriptorInvalid:
		return
	default:
		t.Errorf("Expected an error of type *ErrValueDescriptorInvalid")
	}
}
