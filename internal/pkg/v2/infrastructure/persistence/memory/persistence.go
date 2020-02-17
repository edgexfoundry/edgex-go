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

package memory

import (
	contract "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/persistence/core/metadata"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure/persistence/memory/core/metadata"
)

// Store is the receiver for the metadata in-memory persistence implementation.
type Store struct {
	metadata *metadata.Store
}

// New is a factory function that returns an initialized metadata receiver struct.
func New() *Store {
	return &Store{
		metadata: metadata.New(),
	}
}

// Metadata returns metadata service's persistence implementation.
func (p *Store) Metadata() contract.Store {
	return p.metadata
}
