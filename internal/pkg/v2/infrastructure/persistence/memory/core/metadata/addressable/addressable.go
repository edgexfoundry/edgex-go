/*******************************************************************************
 * Copyright 2020 Dell Inc.
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

package addressable

import (
	"sync"

	model "github.com/edgexfoundry/edgex-go/internal/pkg/v2/domain/models/core/metadata"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure"
)

// data defines the map used to store addressable data.
type data map[infrastructure.Identity]model.Addressable

// Store stores the addressable model in-memory.
type Store struct {
	m    sync.Mutex
	data data
}

// New is a factory function that returns an addressable.
func New() *Store {
	return &Store{
		data: make(data),
	}
}

// NextIdentity returns a new valid identity.
func (_ *Store) NextIdentity() infrastructure.Identity {
	return infrastructure.NewIdentity()
}

// FindByID returns the addressable corresponding to id (if it exists).
func (p *Store) FindByID(id infrastructure.Identity) (*model.Addressable, infrastructure.Status) {
	p.m.Lock()
	defer p.m.Unlock()
	if addressable, exists := p.data[id]; exists {
		return &addressable, infrastructure.StatusSuccess
	}
	return nil, infrastructure.StatusPersistenceNotFound
}

// Save stores addressable.
func (p *Store) Save(addressable model.Addressable) infrastructure.Status {
	p.m.Lock()
	defer p.m.Unlock()
	p.data[addressable.ID] = addressable
	return infrastructure.StatusSuccess
}

// Remove deletes addressable.
func (p *Store) Remove(id infrastructure.Identity) infrastructure.Status {
	p.m.Lock()
	defer p.m.Unlock()
	if _, exists := p.data[id]; exists {
		delete(p.data, id)
		return infrastructure.StatusSuccess
	}
	return infrastructure.StatusPersistenceNotFound
}

// IDOf is a helper that returns the identity of addressable.
func (_ *Store) IDOf(addressable model.Addressable) infrastructure.Identity {
	return addressable.ID
}
