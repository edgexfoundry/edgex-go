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

package mocks

import (
	"os"
	"testing"

	. "github.com/edgexfoundry/edgex-go/internal/security/fileioperformer"
	"github.com/stretchr/testify/assert"
)

func TestMockInterfaceType(t *testing.T) {
	// Typecast will fail if doesn't implement interface properly
	var iface FileIoPerformer = &MockFileIoPerformer{}
	assert.NotNil(t, iface)
}

func TestMockOpenFileReader(t *testing.T) {
	mockFileIoPerformer := &MockFileIoPerformer{}
	mockFileIoPerformer.On("OpenFileReader", "/dev/null", os.O_RDONLY, os.FileMode(0400)).Return(&fakeReaderWriter{}, nil)

	obj, err := mockFileIoPerformer.OpenFileReader("/dev/null", os.O_RDONLY, os.FileMode(0400))
	assert.NotNil(t, obj)
	assert.Nil(t, err)
	mockFileIoPerformer.AssertExpectations(t)
}

func TestMockOpenFileWriter(t *testing.T) {
	mockFileIoPerformer := &MockFileIoPerformer{}
	mockFileIoPerformer.On("OpenFileWriter", "/dev/null", os.O_WRONLY, os.FileMode(0500)).Return(&fakeReaderWriter{}, nil)

	obj, err := mockFileIoPerformer.OpenFileWriter("/dev/null", os.O_WRONLY, os.FileMode(0500))
	assert.NotNil(t, obj)
	assert.Nil(t, err)
	mockFileIoPerformer.AssertExpectations(t)
}

func TestMockMkdirAll(t *testing.T) {
	mockFileIoPerformer := &MockFileIoPerformer{}
	mockFileIoPerformer.On("MkdirAll", "/tmp/foo", os.FileMode(0700)).Return(nil)

	err := mockFileIoPerformer.MkdirAll("/tmp/foo", os.FileMode(0700))
	assert.Nil(t, err)
	mockFileIoPerformer.AssertExpectations(t)
}

//
// fakes
//

type fakeReaderWriter struct{}

func (*fakeReaderWriter) Close() error {
	return nil
}

func (m *fakeReaderWriter) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (m *fakeReaderWriter) Write(p []byte) (n int, err error) {
	return 0, nil
}
