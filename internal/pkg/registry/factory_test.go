//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package registry

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewRegistryClientConsul(t *testing.T) {
	registryInfo := config.RegistryInfo{Type: "consul", Host: "localhost", Port: 8500}
	serviceInfo := config.ServiceInfo{Host: "localhost", Port: 8080}
	serviceKey := "edgex-registry-tests"

	client, err := NewRegistryClient(registryInfo, &serviceInfo, serviceKey)
	if assert.Nil(t, err, "New consul client failed: ", err) == false {
		t.Fatal()
	}

	assert.True(t, client.IsRegistryRunning(), "Consul client should be running")
}
