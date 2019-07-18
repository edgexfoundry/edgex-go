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
package cert

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	skFileExt   = ".priv.key"
	certFileExt = ".pem"
)

// RootCA parameters from JSON: x509_root_ca_parameters
type RootCA struct {
	CAName     string `json:"ca_name"`
	CACountry  string `json:"ca_c"`
	CAState    string `json:"ca_st"`
	CALocality string `json:"ca_l"`
	CAOrg      string `json:"ca_o"`
}

// TLSServer parameters from JSON config: x509_tls_server_parameters
type TLSServer struct {
	TLSHost     string `json:"tls_host"`
	TLSDomain   string `json:"tls_domain"`
	TLSCountry  string `json:"tls_c"`
	TLSSate     string `json:"tls_st"`
	TLSLocality string `json:"tls_l"`
	TLSOrg      string `json:"tls_o"`
}

// X509Config JSON config file main elements
type X509Config struct {
	CreateNewRootCA string    `json:"create_new_rootca"`
	WorkingDir      string    `json:"working_dir"`
	PKISetupDir     string    `json:"pki_setup_dir"`
	RootCA          RootCA    `json:"x509_root_ca_parameters"`
	TLSServer       TLSServer `json:"x509_tls_server_parameters"`
}

// X509ConfigReader is the file reader to read in X509Config JSON format
type X509ConfigReader struct {
	filePath string
}

// NewX509ConfigReader is the constructor to instantiate X509ConfigReader
// returns nil if filepath is empty
func NewX509ConfigReader(filePath string) *X509ConfigReader {
	if filePath == "" {
		return nil
	}
	return &X509ConfigReader{
		filePath: filePath,
	}
}

// Read function reads the config JSON file into X509Config structure
func (reader *X509ConfigReader) Read() (X509Config, error) {
	var jsonX509Config X509Config

	// Open JSON config file
	jsonConfigFile, err := os.Open(reader.filePath)
	defer jsonConfigFile.Close()
	if err != nil {
		return jsonX509Config, err
	}

	// Read JSON config file into byteArray
	byteValue, err := ioutil.ReadAll(jsonConfigFile)
	if err != nil {
		return jsonX509Config, err
	}

	// Initialize the final X509 Configuration array
	// Unmarshal byteArray with the jsonConfigFile's content into jsonX509Config
	err = json.Unmarshal(byteValue, &jsonX509Config)
	if err != nil {
		return jsonX509Config, err
	}

	return jsonX509Config, nil
}

// GetPkiOutputDir returns the dirctory path name derived from X509Config's fields
func (config *X509Config) GetPkiOutputDir() (outputDir string, err error) {
	workingDir, err := filepath.Abs(config.WorkingDir)
	if err != nil {
		return "", err
	}

	return filepath.Join(workingDir, config.PKISetupDir, config.RootCA.CAName), nil
}

// GetCAPemFileName returns the file name of CA certificate
func (config *X509Config) GetCAPemFileName() string {
	return config.RootCA.CAName + certFileExt
}

// GetCAPrivateKeyFileName returns the file name of CA private key
func (config *X509Config) GetCAPrivateKeyFileName() string {
	return config.RootCA.CAName + skFileExt
}

// GetTLSPemFileName returns the file name of TLS certificate
func (config *X509Config) GetTLSPemFileName() string {
	return config.TLSServer.TLSHost + certFileExt
}

// GetTLSPrivateKeyFileName returns the file name of TLS private key
func (config *X509Config) GetTLSPrivateKeyFileName() string {
	return config.TLSServer.TLSHost + skFileExt
}
