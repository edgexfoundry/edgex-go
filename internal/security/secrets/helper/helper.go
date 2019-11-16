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
	envXdgRuntimeDir              = "XDG_RUNTIME_DIR"
	defaultWorkDir                = "/tmp"
	pkiInitBaseDir                = "edgex/security-secrets-setup"
	defaultPkiCacheDir            = "/etc/edgex/pki"
	defaultPkiDeployDir           = "/run/edgex/secrets"
	pkiInitFilePerServiceComplete = ".security-secrets-secrets.complete"
)

type Helper struct {
	loggingClient logger.LoggingClient
	configuration *config.ConfigurationStruct
}

func NewHelper(loggingClient logger.LoggingClient, configuration *config.ConfigurationStruct) *Helper {
	return &Helper{
		loggingClient: loggingClient,
		configuration: configuration,
	}
}

func (h *Helper) CopyFile(fileSrc, fileDest string) (int64, error) {
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

func (h *Helper) CreateDirectoryIfNotExists(dirName string) (err error) {
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err = os.MkdirAll(dirName, os.ModePerm)
	}
	return err
}

func (h *Helper) IsDirEmpty(dir string) (empty bool, err error) {
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

func (h *Helper) CopyDir(srcDir, destDir string) error {
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
			if err := h.CopyDir(srcFilePath, destFilePath); err != nil {
				return err
			}
		} else {
			h.loggingClient.Debug(fmt.Sprintf("copying srcFilePath: %s to destFilePath: %s", srcFilePath, destFilePath))
			if _, copyErr := h.CopyFile(srcFilePath, destFilePath); copyErr != nil {
				return copyErr
			}

		}
	}

	return nil
}

func (h *Helper) Deploy(srcDir, destDir string) error {
	if err := h.CopyDir(srcDir, destDir); err != nil {
		return err
	}
	return h.MarkComplete(destDir)
}

func (h *Helper) SecureEraseFile(fileToErase string) error {
	if err := h.ZeroOutFile(fileToErase); err != nil {
		return err
	}
	return os.Remove(fileToErase)
}

func (h *Helper) ZeroOutFile(fileToErase string) error {
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

func (h *Helper) CheckIfFileExists(fileName string) bool {
	fileInfo, statErr := os.Stat(fileName)
	if os.IsNotExist(statErr) {
		return false
	}
	return !fileInfo.IsDir()
}

func (h *Helper) CheckDirExistsAndIsWritable(path string) (exists bool, isWritable bool) {
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

func (h *Helper) WriteSentinel(sentinelFilename string) error {
	timestamp := []byte(strconv.FormatInt(time.Now().Unix(), 10))
	return ioutil.WriteFile(sentinelFilename, timestamp, 0400)
}

// MarkComplete creates sentinel files of all services to signal pki-init Deploy is done
func (h *Helper) MarkComplete(dirPath string) error {
	// recursively walk through all sub-directories until reach the leaf node
	// to write the sentinel files
	fileDescs, readErr := ioutil.ReadDir(dirPath)
	if readErr != nil {
		return readErr
	}

	for _, fileDesc := range fileDescs {
		aFilePath := filepath.Join(dirPath, fileDesc.Name())

		if fileDesc.IsDir() {
			if err := h.MarkComplete(aFilePath); err != nil {
				return err
			}
		} else {
			// now we are at the leaf node, write sentinel file if not yet
			deployPathDir := path.Dir(aFilePath)
			sentinel := filepath.Join(deployPathDir, pkiInitFilePerServiceComplete)
			if !h.CheckIfFileExists(sentinel) {
				if err := h.WriteSentinel(sentinel); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (h *Helper) GetWorkDir() (string, error) {
	var workDir string
	var err error

	if xdgRuntimeDir, ok := os.LookupEnv(envXdgRuntimeDir); ok {
		workDir = filepath.Join(xdgRuntimeDir, pkiInitBaseDir)
	} else if h.configuration.SecretsSetup.WorkDir != "" {
		workDir, err = filepath.Abs(h.configuration.SecretsSetup.WorkDir)
		if err != nil {
			return "", err
		}
	} else {
		workDir = filepath.Join(defaultWorkDir, pkiInitBaseDir)
	}

	return workDir, nil
}

func (h *Helper) GetCertConfigDir() (string, error) {
	var certConfigDir string
	if h.configuration.SecretsSetup.CertConfigDir != "" {
		certConfigDir = h.configuration.SecretsSetup.CertConfigDir

		// we only read files from CertConfigDir, don't care if it's writable
		if exist, _ := h.CheckDirExistsAndIsWritable(certConfigDir); !exist {
			return "", fmt.Errorf("CertConfigDir from x509 file does not exist in: %s", certConfigDir)
		}

		return certConfigDir, nil
	}

	return "", errors.New("Directory for certificate configuration files not configured")
}

func (h *Helper) GetCacheDir() (string, error) {
	cacheDir := defaultPkiCacheDir

	if h.configuration.SecretsSetup.CacheDir != "" {
		cacheDir = h.configuration.SecretsSetup.CacheDir
		exist, isWritable := h.CheckDirExistsAndIsWritable(cacheDir)
		if !exist {
			return "", fmt.Errorf("CacheDir, %s, from x509 file does not exist", cacheDir)
		}

		if !isWritable {
			return "", fmt.Errorf("CacheDir, %s, from x509 file is not writable", cacheDir)
		}
	}

	return cacheDir, nil
}

func (h *Helper) GetDeployDir() (string, error) {
	deployDir := defaultPkiDeployDir

	if h.configuration.SecretsSetup.DeployDir != "" {
		deployDir = h.configuration.SecretsSetup.DeployDir
		exist, isWritable := h.CheckDirExistsAndIsWritable(deployDir)
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
func (h *Helper) GenTLSAssets(jsonConfig string) error {
	// Read the Json x509 file and unmarshall content into struct type X509Config
	x509Config, err := x509.NewX509Config(jsonConfig)
	if err != nil {
		return err
	}

	directoryHandler := directory.NewHandler(h.loggingClient)
	if directoryHandler == nil {
		return errors.New("h.loggingClient is nil")
	}

	seed, err := seed.NewCertificateSeed(x509Config, directoryHandler)
	if err != nil {
		return err
	}

	rootCA, err := certificates.NewCertificateGenerator(certificates.RootCertificate, seed, certificates.NewFileWriter(), h.loggingClient)
	if err != nil {
		return err
	}

	err = rootCA.Generate()
	if err != nil {
		return err
	}

	tlsCert, err := certificates.NewCertificateGenerator(certificates.TLSCertificate, seed, certificates.NewFileWriter(), h.loggingClient)
	if err != nil {
		return err
	}

	err = tlsCert.Generate()
	if err != nil {
		return err
	}

	return nil
}
