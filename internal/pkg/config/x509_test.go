/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright 2019 Intel Corp.
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

package config

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testJSONFileName = "./testdata/pkisetup-test.json"

func TestLoadX509Config(t *testing.T) {
	tests := []struct {
		name        string
		file        string
		expectError bool
	}{
		{"509LoadOK", testJSONFileName, false},
		{"509NonexistentFile", "./testdata/pkisetup-nope.json", true},
		{"509InvalidJSON", "./testdata/pkisetup-fail.json", true},
	}
	assert := assert.New(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x509Config, err := NewX509Config(tt.file)

			if tt.expectError {
				assert.NotNil(err, "expected error, none returned")
				assert.Empty(x509Config)
			} else {
				assert.Nil(err, "unexpected error %v", err)
				assert.NotEmpty(x509Config)
			}
		})
	}
}

func TestPkiCAOutDir(t *testing.T) {
	x509Config, err := NewX509Config(testJSONFileName)

	assert := assert.New(t)
	assert.Nil(err)
	assert.NotEmpty(x509Config)

	pkiDir, err := x509Config.PkiCADir()
	assert.Nil(err)
	assert.NotEmpty(pkiDir)
	assert.True(strings.Contains(pkiDir,
		filepath.Join(filepath.Base(x509Config.WorkingDir), x509Config.PKISetupDir, x509Config.RootCA.CAName)))
}
