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

package delegate

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure/test"

	"github.com/stretchr/testify/assert"
)

func TestApply(t *testing.T) {
	version := test.FactoryRandomString()
	kind := test.FactoryRandomString()
	action := test.FactoryRandomString()
	before := test.FactoryRandomString()
	expected := test.FactoryRandomString()
	after := test.FactoryRandomString()
	variations := []struct {
		name     string
		handlers []Handler
		expected string
	}{
		{
			name:     test.None,
			handlers: []Handler{},
			expected: expected,
		},
		{
			name: test.One,
			handlers: []Handler{
				newTestHandler(before, after).Handler,
			},
			expected: before + concat(version, kind, action) + expected + after,
		},
	}
	for variationIndex := range variations {
		t.Run(
			test.Name("apply", variations[variationIndex].name),
			func(t *testing.T) {
				routable := Apply(
					application.NewBehavior(version, kind, action),
					TestExecutable,
					variations[variationIndex].handlers,
				)

				response, status := routable.Execute(expected)

				assert.Equal(t, variations[variationIndex].expected, response)
				assert.Equal(t, infrastructure.StatusSuccess, status)
			},
		)
	}
}
