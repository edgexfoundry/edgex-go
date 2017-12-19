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
 * @microservice: core-command-go service
 * @author: Spencer Bull, Dell
 * @version: 0.5.0
 *******************************************************************************/
package main

//
// import (
// 	"net/http"
// 	"testing"
//
// 	"github.com/edgexfoundry/core-clients-go/metadataclients"
// 	"github.com/edgexfoundry/core-domain-go/models"
// )
//
// // Testing with Virtual Devices
// var vdID = "5a174e1be4b009e7b37f0e94"
// var cID = "5a174e1ae4b009e7b37f0e83"
// var body = "{ rpm : 42 }"
// var d = createTestingDevice()
//
// func TestCommandByBadDeviceID(t *testing.T) {
// 	put := true
// 	body, status := commandByDeviceID("badDeviceID", "", "{}", put)
// 	if status != http.StatusForbidden && body != nil {
// 		t.Fail()
// 	}
// }
//
// func TestCommandDeviceLocked(t *testing.T) {
// 	put := true
// 	dc := metadataclients.NewDeviceClient(vdID)
// 	dc.UpdateAdminState(vdID, LOCKED)
// 	body, status := commandByDeviceID(vdID, "", "", put)
// 	if status != http.StatusUnprocessableEntity && body != nil {
// 		t.Fail()
// 	}
// }
//
// // TODO: Fix Command Clien tto not error out if command is down.
// func TestCommandBadCommandID(t *testing.T) {
// 	put := true
// 	// Set Device to unlocked
// 	dc := metadataclients.NewDeviceClient(vdID)
// 	dc.UpdateAdminState(vdID, UNLOCKED)
// 	body, status := commandByDeviceID(vdID, "badCommandID", "{}", put)
// 	if status != http.StatusForbidden && body != nil {
// 		t.Fail()
// 	}
// }
//
// func createTestingDevice() models.Device {
// 	dc := metadataclients.NewDeviceClient(vdID)
// 	// TODO : create Device struct
// 	device := models.Device{}
// 	dc.Add(&device)
// 	return device
// }
