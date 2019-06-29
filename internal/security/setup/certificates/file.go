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
	Write(fileName string, contents []byte, permissions os.FileMode) error
}

// NewFileWriter will instantiate an instance of FileWriter for persisting the given file to disk.
func NewFileWriter() FileWriter {
	return fileWriter{}
}

type fileWriter struct{}

// Write performs the actual write operation based on the parameters supplied
func (w fileWriter) Write(fileName string, contents []byte, permissions os.FileMode) (err error) {
	return ioutil.WriteFile(fileName, contents, permissions)
}

// FileReader is an interface that provides abstraction for reading certificate key files from disk. This abstraction is
// necessary for appropriate mocking of this functionality for unit tests
type FileReader interface {
	Read(fileName string) (data []byte, err error)
}

func NewFileReader() FileReader {
	return fileReader{}
}

type fileReader struct{}

// Read performs the actual read operation based on the parameters supplied
func (r fileReader) Read(fileName string) (data []byte, err error) {
	return ioutil.ReadFile(fileName)
}
