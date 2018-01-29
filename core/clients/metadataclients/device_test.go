/*******************************************************************************
 * Copyright 2017 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 * @microservice: core-clients-go library
 * @author: Ryan Comer, Dell
 * @version: 0.5.0
 *******************************************************************************/
package metadataclients

import (
	"fmt"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"gopkg.in/mgo.v2/bson"
	"os"
	"testing"
)

const (
	deviceUrl        = "http://localhost:48081/api/v1/device"
	addressableUrl   = "http://localhost:48081/api/v1/addressable"
	deviceServiceUrl = "http://localhost:48081/api/v1/deviceservice"
	deviceProfileUrl = "http://localhost:48081/api/v1/deviceprofile"
)

// Test adding a device using the device client
func TestAddDevice(t *testing.T) {
	d := models.Device{
		Addressable:    a,
		AdminState:     "UNLOCKED",
		Name:           "Test name for device",
		OperatingState: "ENABLED",
		Profile:        dp,
		Service:        ds,
	}

	_, err := dc.Add(&d)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
}

var dc DeviceClient
var a models.Addressable
var ds models.DeviceService
var dp models.DeviceProfile

// Main method for the tests
func TestMain(m *testing.M) {
	dc = NewDeviceClient(deviceUrl)
	ac := NewAddressableClient(addressableUrl)
	dsc := NewServiceClient(deviceServiceUrl)
	dpc := NewDeviceProfileClient(deviceProfileUrl)

	a = models.Addressable{
		Address: "http://localhost",
		Name:    "Test Addressable",
		Port:    3000,
	}
	id, err := ac.Add(&a)
	if err != nil {
		fmt.Println("Error posting addressable: " + err.Error())
		return
	}
	a.Id = bson.ObjectIdHex(id)

	ds = models.DeviceService{
		AdminState: "UNLOCKED",
		Service: models.Service{
			Addressable:    a,
			Name:           "Test device service",
			OperatingState: "ENABLED",
		},
	}
	id, err = dsc.Add(&ds)
	if err != nil {
		fmt.Println("Error posting device service: " + err.Error())
		return
	}
	ds.Service.Id = bson.ObjectIdHex(id)

	dp = models.DeviceProfile{
		Manufacturer: "Test manufacturer for device profile",
		Model:        "Test model for device profile",
		Name:         "Test name for device profile",
	}
	id, err = dpc.Add(&dp)
	if err != nil {
		fmt.Println("Error posting new device profile: " + err.Error())
		return
	}
	dp.Id = bson.ObjectIdHex(id)

	os.Exit(m.Run())
}
