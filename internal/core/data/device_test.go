package data

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata/mocks"
	mdMocks "github.com/edgexfoundry/go-mod-core-contracts/clients/metadata/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/edgexfoundry/go-mod-messaging/messaging"
	msgTypes "github.com/edgexfoundry/go-mod-messaging/pkg/types"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
)

var testEvent contract.Event
var testDevice contract.Device
var testDeviceProfile contract.DeviceProfile
var testRoutes *mux.Router

const (
	testDeviceName          string = "Test Device"
	testOrigin              int64  = 123456789
	testBsonString          string = "57e59a71e4b0ca8e6d6d4cc2"
	testUUIDString          string = "ca93c8fa-9919-4ec5-85d3-f81b2b6a7bc1"
	testDeviceProfileName   string = "Test DeviceProfile Name"
	testValueDescriptorName string = "RandomValue_Bool"
)

func TestCheckMaxLimit(t *testing.T) {
	reset()

	testedLimit := math.MinInt32

	expectedNil := checkMaxLimit(testedLimit)

	if expectedNil != nil {
		t.Errorf("Should not exceed limit")
	}
}

func TestCheckMaxLimitOverLimit(t *testing.T) {
	reset()

	testedLimit := math.MaxInt32

	expectedErr := checkMaxLimit(testedLimit)

	if expectedErr == nil {
		t.Errorf("Exceeded limit and should throw error")
	}
}

func TestMain(m *testing.M) {
	testRoutes = LoadRestRoutes()
	LoggingClient = logger.NewMockClient()
	mdc = newMockDeviceClient()
	mdpc = newMockDeviceProfileClient()
	// no need to mock this since it's all in process
	msgClient, _ = messaging.NewMessageClient(msgTypes.MessageBusConfig{
		PublishHost: msgTypes.HostInfo{
			Host:     "*",
			Protocol: "tcp",
			Port:     5563,
		},
		Type: "zero",
	})
	chEvents = make(chan interface{}, 10)
	os.Exit(m.Run())
}

// Supporting methods
// Reset() re-initializes dependencies for each test
func reset() {
	Configuration = &ConfigurationStruct{}
	testEvent.ID = testBsonString
	testEvent.Device = testDeviceName
	testEvent.Origin = testOrigin
	testEvent.Readings = buildReadings()
	testDevice.ProfileName = testDeviceProfileName

}

func newMockDeviceProfileClient() *mdMocks.DeviceProfileClient {
	dpClient := &mdMocks.DeviceProfileClient{}

	testDeviceProfileGenerator := `{"created":1551711642676,"modified":1551711642676,"origin":0,"description":"Example of Device-Virtual","id":"b06e124f-3e46-483d-b18b-fc1f93f835c6","name":"Test DeviceProfile Name","manufacturer":"IOTech","model":"Device-Virtual-01","labels":["device-virtual-example"],"objects":null,"deviceResources":[{"description":"used to decide whether to re-generate a random value","name":"Enable_Randomization","tag":null,"properties":{"value":{"type":"Bool","readWrite":"W","minimum":null,"maximum":null,"defaultValue":"true","size":null,"word":"2","lsb":null,"mask":"0x00","shift":"0","scale":"1.0","offset":"0.0","base":"0","assertion":null,"signed":true,"precision":null},"units":{"type":"String","readWrite":"R","defaultValue":"Random"}},"attributes":null}],"resources":[{"name":"RandomValue_Bool","get":[{"index":null,"operation":"get","object":"RandomValue_Bool","property":null,"parameter":"RandomValue_Bool","resource":null,"secondary":[],"mappings":{}}],"set":[{"index":null,"operation":"set","object":"RandomValue_Bool","property":null,"parameter":"RandomValue_Bool","resource":"RandomValue_Bool","secondary":[],"mappings":{}}]}],"commands":[{"created":1551711642676,"modified":0,"origin":0,"id":"ed0d88b8-e202-4bdd-a941-606b5c7f40d4","name":"RandomValue_Bool","get":{"path":"/api/v1/device/{deviceId}/RandomValue_Bool","responses":[{"code":"200","description":null,"expectedValues":["RandomValue_Bool"]},{"code":"503","description":"service unavailable","expectedValues":[]}]},"put":{"path":"/api/v1/device/{deviceId}/RandomValue_Bool","responses":[{"code":"200","description":null,"expectedValues":[]},{"code":"503","description":"service unavailable","expectedValues":[]}],"parameterNames":["RandomValue_Bool"]}}]}`
	json.Unmarshal([]byte(testDeviceProfileGenerator), &testDeviceProfile)

	dpClient.On("DeviceProfileForName", testDeviceProfileName, context.Background()).Return(testDeviceProfile, nil)

	return dpClient
}

func newMockDeviceClient() *mocks.DeviceClient {
	client := &mocks.DeviceClient{}

	protocols := getProtocols()

	mockDeviceResultFn := func(id string, ctx context.Context) contract.Device {
		if bson.IsObjectIdHex(id) {
			return contract.Device{Id: id, Name: testDeviceName, ProfileName: testDeviceProfileName, Protocols: protocols}
		}
		return contract.Device{ProfileName: testDeviceProfileName}
	}
	client.On("Device", "valid", context.Background()).Return(mockDeviceResultFn, nil)
	client.On("Device", "404", context.Background()).Return(mockDeviceResultFn,
		types.NewErrServiceClient(http.StatusNotFound, []byte{}))
	client.On("Device", mock.Anything, context.Background()).Return(mockDeviceResultFn, fmt.Errorf("some error"))

	mockDeviceForNameResultFn := func(name string, ctx context.Context) contract.Device {
		device := contract.Device{Id: uuid.New().String(), Name: name, ProfileName: testDeviceProfileName, Protocols: protocols}

		return device
	}
	client.On("DeviceForName", testDeviceName, context.Background()).Return(mockDeviceForNameResultFn, nil)
	client.On("DeviceForName", "404", context.Background()).Return(mockDeviceForNameResultFn,
		types.NewErrServiceClient(http.StatusNotFound, []byte{}))
	client.On("DeviceForName", mock.Anything, context.Background()).Return(mockDeviceForNameResultFn,
		fmt.Errorf("some error"))

	return client
}

func getProtocols() map[string]contract.ProtocolProperties {
	p1 := make(map[string]string)
	p1["host"] = "localhost"
	p1["port"] = "1234"
	p1["unitID"] = "1"

	p2 := make(map[string]string)
	p2["serialPort"] = "/dev/USB0"
	p2["baudRate"] = "19200"
	p2["dataBits"] = "8"
	p2["stopBits"] = "1"
	p2["parity"] = "0"
	p2["unitID"] = "2"

	wrap := make(map[string]contract.ProtocolProperties)
	wrap["modbus-ip"] = p1
	wrap["modbus-rtu"] = p2

	return wrap
}

func buildReadings() []contract.Reading {
	ticks := db.MakeTimestamp()
	r1 := contract.Reading{Id: bson.NewObjectId().Hex(),
		Name:     "Temperature",
		Value:    "45",
		Origin:   testOrigin,
		Created:  ticks,
		Modified: ticks,
		Pushed:   ticks,
		Device:   testDeviceName}

	r2 := contract.Reading{Id: bson.NewObjectId().Hex(),
		Name:     "Pressure",
		Value:    "1.01325",
		Origin:   testOrigin,
		Created:  ticks,
		Modified: ticks,
		Pushed:   ticks,
		Device:   testDeviceName}
	readings := []contract.Reading{}
	readings = append(readings, r1, r2)
	return readings
}

func handleDomainEvents(bitEvents []bool, wait *sync.WaitGroup, t *testing.T) {
	until := time.Now().Add(500 * time.Millisecond) //Kill this loop after half second.
	for time.Now().Before(until) {
		select {
		case evt := <-chEvents:
			switch evt.(type) {
			case DeviceLastReported:
				e := evt.(DeviceLastReported)
				if e.DeviceName != testDeviceName {
					t.Errorf("DeviceLastReported name mismatch %s", e.DeviceName)
					return
				}
				bitEvents[0] = true
				break
			case DeviceServiceLastReported:
				e := evt.(DeviceServiceLastReported)
				if e.DeviceName != testDeviceName {
					t.Errorf("DeviceLastReported name mismatch %s", e.DeviceName)
					return
				}
				bitEvents[1] = true
				break
			}
		default:
			//	Without a default case in here, the select block will hang.
		}
	}
	wait.Done()
}
