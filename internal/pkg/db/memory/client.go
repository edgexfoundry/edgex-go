/*******************************************************************************
 * Copyright 2018 Cavium
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
 *******************************************************************************/
package memory

import (
	"github.com/edgexfoundry/edgex-go/internal/export"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

type MemDB struct {
	// Data
	readings     []models.Reading
	events       []models.Event
	vDescriptors []models.ValueDescriptor

	// Metadata
	addressables      []models.Addressable
	commands          []models.Command
	deviceServices    []models.DeviceService
	schedules         []models.Schedule
	scheduleEvents    []models.ScheduleEvent
	provisionWatchers []models.ProvisionWatcher
	deviceReports     []models.DeviceReport
	deviceProfiles    []models.DeviceProfile
	devices           []models.Device

	// Export
	regs []export.Registration
}

func (m *MemDB) CloseSession() {
}

func (m *MemDB) Connect() error {
	return nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

