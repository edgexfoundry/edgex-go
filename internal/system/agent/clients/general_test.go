/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

package clients

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/endpoint"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/general"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"

	"github.com/stretchr/testify/assert"
)

func TestEmptyGetReturnsExpectedValues(t *testing.T) {
	sut := NewGeneral()

	result, ok := sut.Get("nonExistentKey")

	assert.Equal(t, nil, result)
	assert.Equal(t, false, ok)
}

func TestGetForKnownReturnsExpectedValues(t *testing.T) {
	const clientName = "clientName"

	sut := NewGeneral()
	client := general.NewGeneralClient(types.EndpointParams{}, endpoint.Endpoint{RegistryClient: nil})
	sut.Set(clientName, client)

	result, ok := sut.Get(clientName)

	assert.Equal(t, client, result)
	assert.Equal(t, true, ok)
}
