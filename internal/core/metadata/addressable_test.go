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
	"errors"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"os"
	"testing"

	metadataErrors "github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces"
	dbMock "github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/stretchr/testify/mock"
	"gopkg.in/mgo.v2/bson"
)

func TestGetAllAddressables(t *testing.T) {
	reset()
	dbClient = newMockDb()

	const expectedAddressables = 3

	addressables, err := getAllAddressables()
	if err != nil {
		t.Errorf(err.Error())
	}
	if len(addressables) != expectedAddressables {
		t.Errorf("expected addressable count %d, received: %d", expectedAddressables, len(addressables))
	}
}

func TestGetAddressablesFails(t *testing.T) {
	reset()
	dbClient = newMockDb(errors.New("some error"))

	_, expectedErr := getAllAddressables()
	if expectedErr == nil {
		t.Errorf("expected an error from getAllAddressables()")
	}
}

func TestGetAddressablesAboveReadMaxLimit(t *testing.T) {
	reset()
	dbClient = newMockDb()

	Configuration.Service.ReadMaxLimit = 1

	expectedNil, expectedErr := getAllAddressables()

	if expectedNil != nil {
		t.Errorf("getAllAddressables() should return nil when ReadMaxLimit is exceeded")
	}
	if expectedErr == nil {
		t.Errorf("ReadMaxLimit exceeded should return an error")
	}
}

func TestAddAddressable(t *testing.T) {
	reset()
	dbClient = newMockDb()

	objectId := bson.NewObjectId()
	newAddr := models.Addressable{
		Id:   objectId,
		Name: "new addressable",
	}

	id, err := addAddressable(newAddr)

	if err != nil {
		t.Errorf(err.Error())
	}
	if id != objectId.Hex() {
		t.Errorf("Id returned by addAddressable() doesn't match expectation")
	}
}

func TestAddAddressableEmptyName(t *testing.T) {
	reset()
	dbClient = newMockDb()

	objectId := bson.NewObjectId()
	newAddr := models.Addressable{
		Id: objectId,
	}

	_, expectedErr := addAddressable(newAddr)

	if expectedErr == nil {
		t.Errorf("addAddressable() with empty addressable name should cause error")
		return
	}

	if _, ok := expectedErr.(*metadataErrors.ErrEmptyAddressableName); !ok {
		t.Errorf("expected an ErrEmptyAddressableName, found: %s", expectedErr.Error())
	}
}

func TestAddDuplicateAddressableName(t *testing.T) {
	reset()
	dbClient = newMockDb(db.ErrNotUnique)

	objectId := bson.NewObjectId()
	newAddr := models.Addressable{
		Id:   objectId,
		Name: "new addressable",
	}

	_, expectedErr := addAddressable(newAddr)

	if expectedErr == nil {
		t.Errorf("addAddressable() with duplicate addressable name should cause error")
	}

	if _, ok := expectedErr.(*metadataErrors.ErrDuplicateAddressableName); !ok {
		t.Errorf("expected an ErrDuplicateAddressableName, found: %s", expectedErr.Error())
	}
}

func TestMain(m *testing.M) {
	LoggingClient = logger.NewMockClient()
	os.Exit(m.Run())
}

// Supporting methods
// reset() re-initializes dependencies for each test
func reset() {
	Configuration = &ConfigurationStruct{}
	Configuration.Service.ReadMaxLimit = 100
}

func buildAddressables() []models.Addressable {
	a1 := models.Addressable{
		Id:         bson.NewObjectId(),
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
		Id:         bson.NewObjectId(),
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
		Id:         bson.NewObjectId(),
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

	getAddressablesMockFn := func(results *[]models.Addressable) error {
		for _, element := range buildAddressables() {
			*results = append(*results, element)
		}
		return err
	}

	addAddressableMockFn := func(addr *models.Addressable) bson.ObjectId {
		id := addr.Id
		return id
	}

	DB.On("GetAddressables", mock.Anything).Return(getAddressablesMockFn)

	DB.On("AddAddressable", mock.AnythingOfType("*models.Addressable")).Return(addAddressableMockFn, err)

	return DB
}
