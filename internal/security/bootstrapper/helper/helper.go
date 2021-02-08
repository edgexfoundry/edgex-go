/*******************************************************************************
 * Copyright 2021 Intel Corporation
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
 *
 *******************************************************************************/

package helper

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// MarkComplete creates a doneFile file
func MarkComplete(dirPath, doneFile string) error {
	doneFilePath := filepath.Join(dirPath, doneFile)
	if !checkIfFileExists(doneFilePath) {
		if err := writeFile(doneFilePath); err != nil {
			return err
		}
	}

	return nil
}

// CreateDirectoryIfNotExists makes a directory if not exists yet
func CreateDirectoryIfNotExists(dirName string) (err error) {
	if _, err = os.Stat(dirName); err == nil {
		// already exists, skip
		return nil
	} else if os.IsNotExist(err) {
		// dirName not exists yet, create it
		err = os.MkdirAll(dirName, os.ModePerm)
	}

	return
}

func checkIfFileExists(fileName string) bool {
	fileInfo, statErr := os.Stat(fileName)
	if os.IsNotExist(statErr) {
		return false
	}
	return !fileInfo.IsDir()
}

func writeFile(aFileName string) error {
	timestamp := []byte(strconv.FormatInt(time.Now().Unix(), 10))
	return ioutil.WriteFile(aFileName, timestamp, 0400)
}
