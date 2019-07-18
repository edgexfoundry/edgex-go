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
package option

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"
)

func copyFile(fileSrc, fileDest string) (int64, error) {
	var zeroByte int64
	sourceFileSt, err := os.Stat(fileSrc)
	if err != nil {
		return zeroByte, err
	}

	// only regular file allows to be copied
	if !sourceFileSt.Mode().IsRegular() {
		return zeroByte, fmt.Errorf("[%s] is not a regular file to be copied", fileSrc)
	}

	// now open the source file
	source, openErr := os.Open(fileSrc)
	if openErr != nil {
		return zeroByte, openErr
	}
	defer source.Close()

	if _, err := os.Stat(fileDest); err == nil {
		// if fileDest alrady exists, remove it first before create a new one
		os.Remove(fileDest)
	}

	// now create a new file
	dest, createErr := os.Create(fileDest)
	if createErr != nil {
		return zeroByte, createErr
	}
	defer dest.Close()

	bytesWritten, copyErr := io.Copy(dest, source)
	if copyErr != nil {
		return zeroByte, copyErr
	}
	// make dest has the same file mode as the source
	os.Chmod(fileDest, sourceFileSt.Mode())
	return bytesWritten, nil
}

func createDirectoryIfNotExists(dirName string) (err error) {
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err = os.MkdirAll(dirName, os.ModePerm)
	}
	return err
}

func isDirEmpty(dir string) (empty bool, err error) {
	handle, err := os.Open(dir)
	if err != nil {
		return empty, err
	}
	defer handle.Close()

	_, err = handle.Readdir(1)
	if err == io.EOF {
		// EOF error means the dir is empty
		empty = true
		err = nil
	}

	return empty, err
}

func copyDir(srcDir, destDir string) error {
	srcFileInfo, statErr := os.Stat(srcDir)
	if statErr != nil {
		return statErr
	}
	if makeDirErr := os.MkdirAll(destDir, srcFileInfo.Mode()); makeDirErr != nil {
		return makeDirErr
	}

	fileDescs, readErr := ioutil.ReadDir(srcDir)
	if readErr != nil {
		return readErr
	}

	for _, fileDesc := range fileDescs {
		srcFilePath := filepath.Join(srcDir, fileDesc.Name())
		destFilePath := filepath.Join(destDir, fileDesc.Name())

		if fileDesc.IsDir() {
			if err := copyDir(srcFilePath, destFilePath); err != nil {
				return err
			}
		} else {
			log.Printf("copying srcFilePath: %s to destFilePath: %s", srcFilePath, destFilePath)
			if _, copyErr := copyFile(srcFilePath, destFilePath); copyErr != nil {
				return copyErr
			}

		}
	}

	return nil
}

func getPkiCacheDirEnv() string {
	pkiCacheDir := os.Getenv(envPkiCache)
	if pkiCacheDir == "" {
		return defaultPkiCacheDir
	}
	return pkiCacheDir
}

func deploy(srcDir, destDir string) error {
	if err := copyDir(srcDir, destDir); err != nil {
		return err
	}
	return markComplete(destDir)
}

func secureEraseFile(fileToErase string) error {
	if err := zeroOutFile(fileToErase); err != nil {
		return err
	}
	return os.Remove(fileToErase)
}

func zeroOutFile(fileToErase string) error {
	// grant the file read-write permission first
	os.Chmod(fileToErase, 0600)
	fileHdl, err := os.OpenFile(fileToErase, os.O_RDWR, 0600)
	defer fileHdl.Close()
	if err != nil {
		return err
	}
	fileInfo, err := fileHdl.Stat()
	if err != nil {
		return err
	}

	fileSize := fileInfo.Size()
	zeroBytes := make([]byte, fileSize)
	copy(zeroBytes[:], "0")

	// fill the file with zero bytes
	if _, err := fileHdl.Write(zeroBytes); err != nil {
		return err
	}
	return nil
}

func checkIfFileExists(fileName string) bool {
	fileInfo, statErr := os.Stat(fileName)
	if os.IsNotExist(statErr) {
		return false
	}
	return !fileInfo.IsDir()
}

func writeSentinel(sentinelFilename string) error {
	timestamp := []byte(strconv.FormatInt(time.Now().Unix(), 10))
	return ioutil.WriteFile(sentinelFilename, timestamp, 0400)
}

// markComplete creates sentinel files of all services to signal pki-init deploy is done
func markComplete(dirPath string) error {
	// recursively walk through all sub-directories until reach the leaf node
	// to write the sentinel files
	fileDescs, readErr := ioutil.ReadDir(dirPath)
	if readErr != nil {
		return readErr
	}

	for _, fileDesc := range fileDescs {
		aFilePath := filepath.Join(dirPath, fileDesc.Name())

		if fileDesc.IsDir() {
			if err := markComplete(aFilePath); err != nil {
				return err
			}
		} else {
			// now we are at the leaf node, write sentinel file if not yet
			deployPathDir := path.Dir(aFilePath)
			sentinel := filepath.Join(deployPathDir, pkiInitFilePerServiceComplete)
			if !checkIfFileExists(sentinel) {
				if err := writeSentinel(sentinel); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
