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

package secretstoreclient

import (
	"strings"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/stretchr/testify/assert"
)

func TestInsecureRequestor(t *testing.T) {
	mockLogger := logger.MockLogger{}

	caller := NewRequestor(mockLogger).Insecure()
	assert.NotNil(t, caller)
}

func TestSecureRequestor(t *testing.T) {
	mockLogger := logger.MockLogger{}

	caller := NewRequestor(mockLogger).WithTLS(strings.NewReader(""), "str")
	assert.NotNil(t, caller)
}
