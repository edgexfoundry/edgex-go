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

package fileioperformer

import (
	"io"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

//
// Note: these tests do real I/O.
// Filenames are chosen so as to make this I/O harmless
//

func TestFileIoPerformerReader(t *testing.T) {

	fileio := NewDefaultFileIoPerformer()

	if isLinux() {
		reader, err := fileio.OpenFileReader("/dev/null", os.O_RDONLY, 0400)
		assert.Nil(t, err)
		assert.NotNil(t, fileio)
		assert.NotNil(t, reader)
	}
}

func TestFileIoPerformerWriter(t *testing.T) {

	fileio := NewDefaultFileIoPerformer()

	if isLinux() {
		writer, err := fileio.OpenFileWriter("/dev/null", os.O_RDONLY, 0400)
		assert.Nil(t, err)
		assert.NotNil(t, fileio)
		assert.NotNil(t, writer)
	}
}

func TestFileIoPerformerMkdirAll(t *testing.T) {

	fileio := NewDefaultFileIoPerformer()

	if isLinux() {
		err := fileio.MkdirAll("/tmp", 0777)
		assert.Nil(t, err)
	}
}

func TestMakeReadCloserPassThru(t *testing.T) {

	var mockReadCloser io.ReadCloser = &mockReadCloser{}
	var mockReader = mockReadCloser.(io.Reader)

	result := MakeReadCloser(mockReader)
	result2 := MakeReadCloser(mockReadCloser)

	// In all cases, should just get the original object back

	assert.Implements(t, (*io.ReadCloser)(nil), result)
	assert.Implements(t, (*io.ReadCloser)(nil), result2)

	assert.Equal(t, mockReadCloser, result)
	assert.Equal(t, mockReadCloser, result2)
}

func TestMakeReadCloserNopCloser(t *testing.T) {

	var mockReader = &mockReader{}

	result := MakeReadCloser(mockReader)

	// Should get back an object with a NopCloser wrapper

	assert.Implements(t, (*io.ReadCloser)(nil), result)
}

func isLinux() bool {
	return runtime.GOOS == "linux"
}

//
// mocking support
//

type mockReadCloser struct{}

func (rc *mockReadCloser) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (rc *mockReadCloser) Close() error {
	return nil
}

type mockReader struct{}

func (rc *mockReader) Read(p []byte) (n int, err error) {
	return 0, nil
}
