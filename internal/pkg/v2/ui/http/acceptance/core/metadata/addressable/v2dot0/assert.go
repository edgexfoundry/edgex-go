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

package v2dot0

import (
	"testing"

	dtoCreateV2dot0 "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/core/metadata/addressable/create"
	dtoReadV2dot0 "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/core/metadata/addressable/read"

	"github.com/stretchr/testify/assert"
)

// AssertCreated asserts the addressable in a readResponse equals the addressable in an createRequest.
func AssertCreated(t *testing.T, createRequest *dtoCreateV2dot0.Request, readResponse *dtoReadV2dot0.Response) {
	assert.Equal(t, createRequest.Name, readResponse.Name)
	assert.Equal(t, createRequest.Protocol, readResponse.Protocol)
	assert.Equal(t, createRequest.Method, readResponse.Method)
	assert.Equal(t, createRequest.Address, readResponse.Address)
	assert.Equal(t, createRequest.Port, readResponse.Port)
	assert.Equal(t, createRequest.Path, readResponse.Path)
	assert.Equal(t, createRequest.Publisher, readResponse.Publisher)
	assert.Equal(t, createRequest.User, readResponse.User)
	assert.Equal(t, createRequest.Password, readResponse.Password)
	assert.Equal(t, createRequest.Topic, readResponse.Topic)
}
