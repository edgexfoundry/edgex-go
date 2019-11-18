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

package helper

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	x509 "github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/certificates"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/directory"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/seed"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

const (
	EnvXdgRuntimeDir              = "XDG_RUNTIME_DIR"
	defaultWorkDir                = "/tmp"
	pkiInitBaseDir                = "edgex/security-secrets-setup"
	defaultPkiCacheDir            = "/etc/edgex/pki"
	defaultPkiDeployDir           = "/run/edgex/secrets"
	PkiInitFilePerServiceComplete = ".security-secrets-secrets.complete"
)

func CopyFile(fileSrc, fileDest string) (int64, error) {
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
	_ = os.Chmod(fileDest, sourceFileSt.Mode())
	return bytesWritten, nil
}

func CreateDirectoryIfNotExists(dirName string) (err error) {
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err = os.MkdirAll(dirName, os.ModePerm)
	}
	return err
}

func IsDirEmpty(dir string) (empty bool, err error) {
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

func CopyDir(srcDir, destDir string, loggingClient logger.LoggingClient) error {
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
			if err := CopyDir(srcFilePath, destFilePath, loggingClient); err != nil {
				return err
			}
		} else {
			loggingClient.Debug(fmt.Sprintf("copying srcFilePath: %s to destFilePath: %s", srcFilePath, destFilePath))
			if _, copyErr := CopyFile(srcFilePath, destFilePath); copyErr != nil {
				return copyErr
			}

		}
	}

	return nil
}

func Deploy(srcDir, destDir string, loggingClient logger.LoggingClient) error {
	if err := CopyDir(srcDir, destDir, loggingClient); err != nil {
		return err
	}
	return markComplete(destDir)
}

func SecureEraseFile(fileToErase string) error {
	if err := zeroOutFile(fileToErase); err != nil {
		return err
	}
	return os.Remove(fileToErase)
}

func zeroOutFile(fileToErase string) error {
	// grant the file read-write permission first
	_ = os.Chmod(fileToErase, 0600)
	fileHdl, err := os.OpenFile(fileToErase, os.O_RDWR, 0600)
	defer func() {
		_ = fileHdl.Close()
	}()
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

func CheckIfFileExists(fileName string) bool {
	fileInfo, statErr := os.Stat(fileName)
	if os.IsNotExist(statErr) {
		return false
	}
	return !fileInfo.IsDir()
}

func CheckDirExistsAndIsWritable(path string) (exists bool, isWritable bool) {
	exists = false
	isWritable = false

	fileInfo, statErr := os.Stat(path)
	if os.IsNotExist(statErr) {
		return
	}

	exists = fileInfo.IsDir()

	fileMode := uint32(fileInfo.Mode())
	writableMask := uint32(1 << 7)
	if fileMode&writableMask != 0 {
		isWritable = true
	}
	return
}

func writeSentinel(sentinelFilename string) error {
	timestamp := []byte(strconv.FormatInt(time.Now().Unix(), 10))
	return ioutil.WriteFile(sentinelFilename, timestamp, 0400)
}

// markComplete creates sentinel files of all services to signal pki-init Deploy is done
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
			sentinel := filepath.Join(deployPathDir, PkiInitFilePerServiceComplete)
			if !CheckIfFileExists(sentinel) {
				if err := writeSentinel(sentinel); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func GetWorkDir(configuration *config.ConfigurationStruct) (string, error) {
	var workDir string
	var err error

	if xdgRuntimeDir, ok := os.LookupEnv(EnvXdgRuntimeDir); ok {
		workDir = filepath.Join(xdgRuntimeDir, pkiInitBaseDir)
	} else if configuration.SecretsSetup.WorkDir != "" {
		workDir, err = filepath.Abs(configuration.SecretsSetup.WorkDir)
		if err != nil {
			return "", err
		}
	} else {
		workDir = filepath.Join(defaultWorkDir, pkiInitBaseDir)
	}

	return workDir, nil
}

const errMessageCertConfigDirNotSet = "Directory for certificate configuration files not configured"

func GetCertConfigDir(configuration *config.ConfigurationStruct) (string, error) {
	var certConfigDir string
	if configuration.SecretsSetup.CertConfigDir != "" {
		certConfigDir = configuration.SecretsSetup.CertConfigDir

		// we only read files from CertConfigDir, don't care if it's writable
		if exist, _ := CheckDirExistsAndIsWritable(certConfigDir); !exist {
			return "", fmt.Errorf("CertConfigDir from x509 file does not exist in: %s", certConfigDir)
		}

		return certConfigDir, nil
	}

	return "", errors.New(errMessageCertConfigDirNotSet)
}

func GetCacheDir(configuration *config.ConfigurationStruct) (string, error) {
	cacheDir := defaultPkiCacheDir

	if configuration.SecretsSetup.CacheDir != "" {
		cacheDir = configuration.SecretsSetup.CacheDir
		exist, isWritable := CheckDirExistsAndIsWritable(cacheDir)
		if !exist {
			return "", fmt.Errorf("CacheDir, %s, from x509 file does not exist", cacheDir)
		}

		if !isWritable {
			return "", fmt.Errorf("CacheDir, %s, from x509 file is not writable", cacheDir)
		}
	}

	return cacheDir, nil
}

func GetDeployDir(configuration *config.ConfigurationStruct) (string, error) {
	deployDir := defaultPkiDeployDir

	if configuration.SecretsSetup.DeployDir != "" {
		deployDir = configuration.SecretsSetup.DeployDir
		exist, isWritable := CheckDirExistsAndIsWritable(deployDir)
		if !exist {
			return "", fmt.Errorf("DeployDir, %s, from x509 file does not exist", deployDir)
		}
		if !isWritable {
			return "", fmt.Errorf("DeployDir, %s, from x509 file is not writable", deployDir)
		}
	}

	return deployDir, nil
}

// GenTLSAssets generates the TLS assets based on the JSON configuration file
func GenTLSAssets(jsonConfig string, loggingClient logger.LoggingClient) error {
	// Read the Json x509 file and unmarshall content into struct type X509Config
	x509Config, err := x509.NewX509Config(jsonConfig)
	if err != nil {
		return err
	}

	directoryHandler := directory.NewHandler(loggingClient)
	if directoryHandler == nil {
		return errors.New("h.loggingClient is nil")
	}

	seed, err := seed.NewCertificateSeed(x509Config, directoryHandler)
	if err != nil {
		return err
	}

	rootCA, err := certificates.NewCertificateGenerator(certificates.RootCertificate, seed, certificates.NewFileWriter(), loggingClient)
	if err != nil {
		return err
	}

	err = rootCA.Generate()
	if err != nil {
		return err
	}

	tlsCert, err := certificates.NewCertificateGenerator(certificates.TLSCertificate, seed, certificates.NewFileWriter(), loggingClient)
	if err != nil {
		return err
	}

	err = tlsCert.Generate()
	if err != nil {
		return err
	}

	return nil
}
