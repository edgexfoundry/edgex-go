/********************************************************************************
 *  Copyright 2020 Dell Inc.
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
	"reflect"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/google/uuid"
)

// allTestAddressables a slice containing all the pre-constructed addressables for testing.
var allTestAddressables = []contract.Addressable{addressable1, addressable2, addressable3}

// addressable1 is a pre-constructed addressable with unique properties which does not match any other addressable when
//  invoking functions which can return multiple addressables such as GetAddressablesByPort
var addressable1 = contract.Addressable{
	Timestamps: contract.Timestamps{
		Created:  123,
		Modified: 456,
		Origin:   789,
	},
	// This is a valid UUID which allows us to ensure that the ID used is not replaced.
	Id:         "6d40520f-035b-4be9-bacc-152d4f4f37a9",
	Name:       "FirstAddressable",
	Protocol:   "tcp",
	HTTPMethod: "POST",
	Address:    "0.0.0.0",
	Port:       8000,
	Path:       "/",
	Publisher:  "Mushroom Kingdom",
	User:       "Yoshi",
	Password:   "MarioIsGettingHeavy",
	Topic:      "dinosaur-things",
}

// addressable2 is a pre-constructed addressable with non-unique properties. The properties which can lead to multiple
// results when invoking the client, such as GetAddressablesByPort, have matching properties with addressable3. This
// helps with testing situations where multiple addressables are returned
var addressable2 = contract.Addressable{
	Timestamps: contract.Timestamps{
		Created:  789,
		Modified: 456,
		Origin:   123,
	},
	// This is a valid UUID which allows us to ensure that the ID used is not replaced.
	Id:         "1d874125-df04-4718-bd81-b550019b74aa",
	Name:       "SecondAddressable",
	Protocol:   "http",
	HTTPMethod: "GET",
	Address:    "127.0.0.1",
	Port:       8080,
	Path:       "/home",
	Publisher:  "Nintendo",
	User:       "Link the hero time",
	Password:   "RescueZelda",
	Topic:      "hyrule-events",
}

// addressable3 is a pre-constructed addressable with non-unique properties. The properties which can lead to multiple
// results when invoking the client, such as GetAddressablesByPort, have matching properties with addressable2. This
// helps with testing situations where multiple addressables are returned
var addressable3 = contract.Addressable{
	Timestamps: contract.Timestamps{
		Created:  789,
		Modified: 456,
		Origin:   123,
	},
	// This is a valid UUID which allows us to ensure that the ID used is not replaced.
	Id:         "cb7d56ca-527d-448b-91f0-bc27f9891b90",
	Name:       "ThirdAddressable",
	Protocol:   "stomp",
	HTTPMethod: "PUT",
	Address:    "127.0.0.1",
	Port:       8080,
	Path:       "/castle",
	Publisher:  "Nintendo",
	User:       "Gannon",
	Password:   "GetTriforce",
	Topic:      "hyrule-events",
}

func TestClient_AddAddressableUUID(t *testing.T) {
	client := NewClient()
	generatedId, err := client.AddAddressable(contract.Addressable{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	// Verify the ID is not blank and is a valid UUID
	_, err = uuid.Parse(generatedId)
	if err != nil {
		t.Errorf("The generated UUID is not valid: %s", generatedId)
	}
}

func TestClient_GetAddressables(t *testing.T) {

	tests := []struct {
		name                     string
		prePopulatedAddressables []contract.Addressable
		want                     []contract.Addressable
		wantErr                  bool
	}{
		{
			"One Addressable",
			[]contract.Addressable{addressable1},
			[]contract.Addressable{addressable1},
			false,
		},
		{
			"Multiple Addressables",
			[]contract.Addressable{addressable1, addressable2},

			[]contract.Addressable{addressable1, addressable2},
			false,
		},
		{
			"No Addressables",
			[]contract.Addressable{},

			[]contract.Addressable{},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := createAndPopulateClient(tt.prePopulatedAddressables)
			if err != nil {
				t.Errorf("Unexpected error during test setup: %v", err)
				return
			}

			got, err := c.GetAddressables()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAddressables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !verifySliceEqualIgnoreOrder(got, tt.want) {
				t.Errorf("GetAddressables() got = %v, want %v", got, tt.want)

			}
		})
	}
}

func TestClient_UpdateAddressable(t *testing.T) {

	tests := []struct {
		name                     string
		prePopulatedAddressables []contract.Addressable
		newAddressable           contract.Addressable
		wantErr                  bool
		errorType                error
	}{
		{
			"Update existing addressable",
			allTestAddressables,
			addressable1,
			false,
			nil,
		},
		{
			"Update non-existing addressable",
			[]contract.Addressable{addressable1},
			addressable2,
			true,
			db.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := createAndPopulateClient(tt.prePopulatedAddressables)
			if err != nil {
				t.Errorf("Unexpected error during test setup: %v", err)
				return
			}

			err = c.UpdateAddressable(tt.newAddressable)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateAddressable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errorType != nil {
				if err.Error() != tt.errorType.Error() {
					t.Errorf("Expected error of '%v', but got an error of  '%v'", tt.errorType, err)
					return
				}
			}

			if !tt.wantErr {
				want := tt.newAddressable
				got, err := c.GetAddressableById(tt.newAddressable.Id)
				if err != nil {
					t.Errorf("Unexpected error during validation of update: %v", err)
					return
				}

				if !reflect.DeepEqual(want, got) {
					t.Errorf("UpdateAddressable() got = %v, want %v", got, want)
					return
				}
			}
		})
	}
}

func TestClient_GetAddressableById(t *testing.T) {

	tests := []struct {
		name                     string
		prePopulatedAddressables []contract.Addressable
		id                       string
		want                     contract.Addressable
		wantErr                  bool
		errorType                error
	}{
		{
			"Get existing addressable by ID",
			[]contract.Addressable{addressable1},
			addressable1.Id,
			addressable1,
			false,
			nil,
		},
		{
			"No matching ID",
			[]contract.Addressable{addressable1},
			addressable2.Id,
			contract.Addressable{},
			true,
			db.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := createAndPopulateClient(tt.prePopulatedAddressables)
			if err != nil {
				t.Errorf("Unexpected error during test setup: %v", err)
				return
			}

			got, err := c.GetAddressableById(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAddressableById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errorType != nil {
				if err.Error() != tt.errorType.Error() {
					t.Errorf("Expected error of '%v', but got an error of  '%v'", tt.errorType, err)
					return
				}
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAddressableById() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_AddAddressable(t *testing.T) {
	tests := []struct {
		name                     string
		prePopulatedAddressables []contract.Addressable
		newAddressable           contract.Addressable
		want                     string
		wantErr                  bool
		errorType                error
	}{
		{
			"Add addressable with unique ID",
			[]contract.Addressable{addressable1},
			addressable2,
			addressable2.Id,
			false,
			nil,
		},
		{
			"Add addressable with non-unique ID",
			[]contract.Addressable{addressable1},
			addressable1,
			"",
			true,
			db.ErrNotUnique,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := createAndPopulateClient(tt.prePopulatedAddressables)
			if err != nil {
				t.Errorf("Unexpected error during test setup: %v", err)
				return
			}

			got, err := c.AddAddressable(tt.newAddressable)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddAddressable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errorType != nil {
				if err.Error() != tt.errorType.Error() {
					t.Errorf("Expected error of '%v', but got an error of  '%v'", tt.errorType, err)
					return
				}
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("AddAddressable() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetAddressableByName(t *testing.T) {

	tests := []struct {
		name                     string
		prePopulatedAddressables []contract.Addressable
		addressableName          string
		want                     contract.Addressable
		wantErr                  bool
		errorType                error
	}{
		{
			"Get matching addressable by name",
			allTestAddressables,
			addressable1.Name,
			addressable1,
			false,
			nil,
		},
		{
			"No matching addressable",
			allTestAddressables,
			"non-existent name",
			contract.Addressable{},
			true,
			db.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := createAndPopulateClient(tt.prePopulatedAddressables)
			if err != nil {
				t.Errorf("Unexpected error during test setup: %v", err)
				return
			}

			got, err := c.GetAddressableByName(tt.addressableName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAddressableByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errorType != nil {
				if err.Error() != tt.errorType.Error() {
					t.Errorf("Expected error of '%v', but got an error of  '%v'", tt.errorType, err)
					return
				}
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAddressableByName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetAddressablesByTopic(t *testing.T) {
	tests := []struct {
		name                     string
		prePopulatedAddressables []contract.Addressable
		topic                    string
		want                     []contract.Addressable
		wantErr                  bool
		errorType                error
	}{
		{
			"Get matching addressable by topic",
			allTestAddressables,
			addressable1.Topic,
			[]contract.Addressable{addressable1},
			false,
			nil,
		},
		{
			"Get multiple matching addressable by topic",
			allTestAddressables,
			addressable2.Topic,
			[]contract.Addressable{addressable2, addressable3},
			false,
			nil,
		},
		{
			"No matching addressable",
			allTestAddressables,
			"non-existent topic",
			nil,
			true,
			db.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := createAndPopulateClient(tt.prePopulatedAddressables)
			if err != nil {
				t.Errorf("Unexpected error during test setup: %v", err)
				return
			}

			got, err := c.GetAddressablesByTopic(tt.topic)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAddressablesByTopic() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errorType != nil {
				if err.Error() != tt.errorType.Error() {
					t.Errorf("Expected error of '%v', but got an error of  '%v'", tt.errorType, err)
					return
				}
			}

			if !tt.wantErr && !verifySliceEqualIgnoreOrder(got, tt.want) {
				t.Errorf("GetAddressablesByTopic() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetAddressablesByPort(t *testing.T) {
	tests := []struct {
		name                     string
		prePopulatedAddressables []contract.Addressable
		port                     int
		want                     []contract.Addressable
		wantErr                  bool
		errorType                error
	}{
		{
			"Get matching addressable by port",
			allTestAddressables,
			addressable1.Port,
			[]contract.Addressable{addressable1},
			false,
			nil,
		},
		{
			"Get multiple matching addressable by port",
			allTestAddressables,
			addressable2.Port,
			[]contract.Addressable{addressable2, addressable3},
			false,
			nil,
		},
		{
			"No matching addressable",
			allTestAddressables,
			999999,
			nil,
			true,
			db.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := createAndPopulateClient(tt.prePopulatedAddressables)
			if err != nil {
				t.Errorf("Unexpected error during test setup: %v", err)
				return
			}

			got, err := c.GetAddressablesByPort(tt.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAddressablesByPort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errorType != nil {
				if err.Error() != tt.errorType.Error() {
					t.Errorf("Expected error of '%v', but got an error of  '%v'", tt.errorType, err)
					return
				}
			}

			if !tt.wantErr && !verifySliceEqualIgnoreOrder(got, tt.want) {
				t.Errorf("GetAddressablesByPort() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetAddressablesByAddress(t *testing.T) {
	tests := []struct {
		name                     string
		prePopulatedAddressables []contract.Addressable
		address                  string
		want                     []contract.Addressable
		wantErr                  bool
		errorType                error
	}{
		{
			"Get matching addressable by address",
			allTestAddressables,
			addressable1.Address,
			[]contract.Addressable{addressable1},
			false,
			nil,
		},
		{
			"Get multiple matching addressable by address",
			allTestAddressables,
			addressable2.Address,
			[]contract.Addressable{addressable2, addressable3},
			false,
			nil,
		},
		{
			"No matching addressable",
			allTestAddressables,
			"non-existent address",
			nil,
			true,
			db.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := createAndPopulateClient(tt.prePopulatedAddressables)
			if err != nil {
				t.Errorf("Unexpected error during test setup: %v", err)
				return
			}

			got, err := c.GetAddressablesByAddress(tt.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAddressablesByAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errorType != nil {
				if err.Error() != tt.errorType.Error() {
					t.Errorf("Expected error of '%v', but got an error of  '%v'", tt.errorType, err)
					return
				}
			}

			if !tt.wantErr && !verifySliceEqualIgnoreOrder(got, tt.want) {
				t.Errorf("GetAddressablesByAddress() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetAddressablesByPublisher(t *testing.T) {
	tests := []struct {
		name                     string
		prePopulatedAddressables []contract.Addressable
		publisher                string
		want                     []contract.Addressable
		wantErr                  bool
		errorType                error
	}{
		{
			"Get matching addressable by publisher",
			allTestAddressables,
			addressable1.Publisher,
			[]contract.Addressable{addressable1},
			false,
			nil,
		},
		{
			"Get multiple matching addressable by publisher",
			allTestAddressables,
			addressable2.Publisher,
			[]contract.Addressable{addressable2, addressable3},
			false,
			nil,
		},
		{
			"No matching addressable",
			allTestAddressables,
			"non-existent publisher",
			nil,
			true,
			db.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := createAndPopulateClient(tt.prePopulatedAddressables)
			if err != nil {
				t.Errorf("Unexpected error during test setup: %v", err)
				return
			}

			got, err := c.GetAddressablesByPublisher(tt.publisher)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAddressablesByPublisher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errorType != nil {
				if err.Error() != tt.errorType.Error() {
					t.Errorf("Expected error of '%v', but got an error of  '%v'", tt.errorType, err)
					return
				}
			}

			if !tt.wantErr && !verifySliceEqualIgnoreOrder(got, tt.want) {
				t.Errorf("GetAddressablesByPublisher() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_DeleteAddressableById(t *testing.T) {
	tests := []struct {
		name                     string
		prePopulatedAddressables []contract.Addressable
		id                       string
		wantErr                  bool
		errorType                error
	}{
		{
			"Delete existing addressable",
			allTestAddressables,
			addressable1.Id,
			false,
			nil,
		},
		{
			"Delete non-existing addressable",
			allTestAddressables,
			"non-existing ID",
			true,
			db.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := createAndPopulateClient(tt.prePopulatedAddressables)
			if err != nil {
				t.Errorf("Unexpected error during test setup: %v", err)
				return
			}

			err = c.DeleteAddressableById(tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteAddressableById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errorType != nil {
				if err.Error() != tt.errorType.Error() {
					t.Errorf("Expected error of '%v', but got an error of  '%v'", tt.errorType, err)
					return
				}
			}

			if !tt.wantErr {
				_, err := c.GetAddressableById(tt.id)
				if err == nil || err.Error() != db.ErrNotFound.Error() {
					t.Errorf("Expected the Addressable to be deleted but was not.")
				}
			}
		})
	}
}

// createAndPopulateClient constructs a new client and loads provided addressables into the underlying datastore.
func createAndPopulateClient(addressables []contract.Addressable) (*Client, error) {
	client := NewClient()

	for _, addressable := range addressables {
		_, err := client.AddAddressable(addressable)
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}

// verifySliceEqualIgnoreOrder verfies the contents of two slices of Addressables are equal ignoring order.
func verifySliceEqualIgnoreOrder(a, b []contract.Addressable) bool {
	aLen := len(a)
	bLen := len(b)

	if aLen != bLen {
		return false
	}

	visited := make([]bool, bLen)

	for i := 0; i < aLen; i++ {
		found := false
		element := a[i]
		for j := 0; j < bLen; j++ {
			if visited[j] {
				continue
			}
			if element == b[j] {
				visited[j] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
