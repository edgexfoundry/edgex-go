/*******************************************************************************
 * Copyright 2018 Dell Inc.
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

package metadata

import (
	"os"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces"
	dbMock "github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func TestMain(m *testing.M) {
	LoggingClient = logger.NewMockClient()
	os.Exit(m.Run())
}

// Supporting methods
// reset() re-initializes dependencies for each test
func reset() {
	Configuration = &ConfigurationStruct{}
	Configuration.Service.MaxResultCount = 50000
}

func buildAddressables() []models.Addressable {
	a1 := models.Addressable{
		Id:         uuid.New().String(),
		Name:       "camera address",
		Protocol:   "HTTP",
		HTTPMethod: "get",
		Address:    "172.17.0.1",
		Port:       49999,
		Path:       "/foo",
		Publisher:  "none",
		User:       "none",
		Password:   "none",
		Topic:      "none",
	}
	a2 := models.Addressable{
		Id:         uuid.New().String(),
		Name:       "microphone address",
		Protocol:   "HTTP",
		HTTPMethod: "get",
		Address:    "172.17.0.2",
		Port:       49999,
		Path:       "/foo",
		Publisher:  "none",
		User:       "none",
		Password:   "none",
		Topic:      "none",
	}
	a3 := models.Addressable{
		Id:         uuid.New().String(),
		Name:       "pressure sensor address",
		Protocol:   "HTTP",
		HTTPMethod: "get",
		Address:    "172.17.0.3",
		Port:       49999,
		Path:       "/foo",
		Publisher:  "none",
		User:       "none",
		Password:   "none",
		Topic:      "none",
	}

	addressables := make([]models.Addressable, 0)
	addressables = append(addressables, a1, a2, a3)
	return addressables
}

func newMockDb(errors ...error) interfaces.DBClient {
	DB := &dbMock.DBClient{}

	var err error = nil
	if len(errors) > 0 {
		err = errors[0]
	}

	getAddressablesMockFn := func() []models.Addressable {
		results := make([]models.Addressable, 0)
		for _, element := range buildAddressables() {
			results = append(results, element)
		}
		return results
	}

	addAddressableMockFn := func(addr models.Addressable) string {
		id := addr.Id
		return id
	}

	DB.On("GetAddressables").Return(getAddressablesMockFn, err)

	DB.On("AddAddressable", mock.AnythingOfType("models.Addressable")).Return(addAddressableMockFn, err)

	return DB
}
