//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
// in compliance with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under
// the License.
//
// SPDX-License-Identifier: Apache-2.0'
//

package authtokenloader

import (
	"errors"
	"os"
	"strings"
	"testing"

	. "github.com/edgexfoundry/edgex-go/internal/security/fileioperformer/mocks"
	"github.com/stretchr/testify/assert"
)

const createTokenJSON = `{"auth":{"client_token":"some-token-value"}}`
const vaultInitJSON = `{"root_token":"some-token-value"}`
const expectedToken = "some-token-value"

func TestReadCreateTokenJSON(t *testing.T) {
	stringReader := strings.NewReader(createTokenJSON)
	mockFileIoPerformer := &MockFileIoPerformer{}
	mockFileIoPerformer.On("OpenFileReader", "/dev/null", os.O_RDONLY, os.FileMode(0400)).Return(stringReader, nil)

	p := NewAuthTokenLoader(mockFileIoPerformer)

	token, err := p.Load("/dev/null")
	assert.Nil(t, err)
	assert.Equal(t, expectedToken, token)
}

func TestReadVaultInitJSON(t *testing.T) {
	stringReader := strings.NewReader(vaultInitJSON)
	mockFileIoPerformer := &MockFileIoPerformer{}
	mockFileIoPerformer.On("OpenFileReader", "/dev/null", os.O_RDONLY, os.FileMode(0400)).Return(stringReader, nil)

	p := NewAuthTokenLoader(mockFileIoPerformer)

	token, err := p.Load("/dev/null")
	assert.Nil(t, err)
	assert.Equal(t, expectedToken, token)
}

func TestReadEmptyJSON(t *testing.T) {
	stringReader := strings.NewReader("{}")
	mockFileIoPerformer := &MockFileIoPerformer{}
	mockFileIoPerformer.On("OpenFileReader", "/dev/null", os.O_RDONLY, os.FileMode(0400)).Return(stringReader, nil)

	p := NewAuthTokenLoader(mockFileIoPerformer)

	_, err := p.Load("/dev/null")
	assert.EqualError(t, err, "Unable to find authentication token in /dev/null")
}

func TestFailOpen(t *testing.T) {
	stringReader := strings.NewReader("")
	myerr := errors.New("error")
	mockFileIoPerformer := &MockFileIoPerformer{}
	mockFileIoPerformer.On("OpenFileReader", "/dev/null", os.O_RDONLY, os.FileMode(0400)).Return(stringReader, myerr)

	p := NewAuthTokenLoader(mockFileIoPerformer)

	_, err := p.Load("/dev/null")
	assert.Equal(t, myerr, err)
}
