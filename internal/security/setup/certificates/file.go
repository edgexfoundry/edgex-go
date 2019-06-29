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

package certificates

import (
	"io/ioutil"
	"os"
)

// FileWriter is an interface that provides abstraction for writing certificate files to disk. This abstraction is
// necessary for appropriate mocking of this functionality for unit tests
type FileWriter interface {
	Write() error
}

// NewFileWriter will instantiate an instance of FileWriter for persisting the given file to disk.
func NewFileWriter(fileName string, contents []byte, permissions os.FileMode) FileWriter {
	return fileWriter{name: fileName, data: contents, perms: permissions}
}

type fileWriter struct {
	name  string
	data  []byte
	perms os.FileMode
}

// Write performs the actual write operation based on the parameters supplied to NewFileWriter
func (w fileWriter) Write() (err error) {
	return ioutil.WriteFile(w.name, w.data, w.perms)
}
