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

package metadata

import (
	model "github.com/edgexfoundry/edgex-go/internal/pkg/v2/domain/models/core/metadata"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure"
)

// Addressable defines the repository contract for the addressable model.
type Addressable interface {
	FindByID(id infrastructure.Identity) (*model.Addressable, infrastructure.Status)
	NextIdentity() infrastructure.Identity
	Save(addressable model.Addressable) infrastructure.Status
	Remove(id infrastructure.Identity) infrastructure.Status
	IDOf(addressable model.Addressable) infrastructure.Identity
}
