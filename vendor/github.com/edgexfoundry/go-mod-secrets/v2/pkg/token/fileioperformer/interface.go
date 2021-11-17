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
)

type FileIoPerformer interface {
	// OpenFileReader is a function that opens a file and returns an io.Reader (at a minimum)
	OpenFileReader(name string, flag int, perm os.FileMode) (io.Reader, error)
	// OpenFileWriter is a function that opens a file and returns an io.WriteCloser (at a minimum)
	OpenFileWriter(name string, flag int, perm os.FileMode) (io.WriteCloser, error)
	// MkdirAll creates a directory tree (see os.MkdirAll)
	MkdirAll(path string, perm os.FileMode) error
}
