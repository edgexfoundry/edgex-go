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

package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// NewX509Config will populate a struct representing configuration for X.509 key generation
func NewX509Config(configFilePtr string) (X509Config, error) {

	var jsonX509Config X509Config

	// Open JSON config file
	jsonFile, err := os.Open(configFilePtr)
	if err != nil {
		return jsonX509Config, err
	}
	defer jsonFile.Close()

	// Read JSON config file into byteArray
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return jsonX509Config, err
	}

	// Initialize the final X509 Configuration array
	// Unmarshal byteArray with the jsonFile's content into jsonX509Config
	err = json.Unmarshal(byteValue, &jsonX509Config)
	if err != nil {
		return jsonX509Config, err
	}

	return jsonX509Config, nil
}

// X509Config JSON config file main structure
type X509Config struct {
	CreateNewRootCA string    `json:"create_new_rootca"`
	WorkingDir      string    `json:"working_dir"`
	PKISetupDir     string    `json:"pki_setup_dir"`
	DumpConfig      string    `json:"dump_config"`
	KeyScheme       KeyScheme `json:"key_scheme"`
	RootCA          RootCA    `json:"x509_root_ca_parameters"`
	TLSServer       TLSServer `json:"x509_tls_server_parameters"`
}

func (cfg X509Config) PkiCADir() (string, error) {
	dir, err := filepath.Abs(cfg.WorkingDir)
	if err != nil {
		// Looking at the implementation of filepath.Abs it does NOT verify the existence of the path
		return "", fmt.Errorf("unable to determine absolute path -- %s", err.Error())
	}
	// pkiCaDir: Concatenate working dir absolute path with PKI setup dir, using separator "/"
	return strings.Join([]string{dir, cfg.PKISetupDir, cfg.RootCA.CAName}, "/"), nil
}

// KeyScheme parameters (RSA vs EC)
// RSA: 1024, 2048, 4096
// EC: 224, 256, 384, 521
type KeyScheme struct {
	DumpKeys   string `json:"dump_keys"`
	RSA        string `json:"rsa"`
	RSAKeySize string `json:"rsa_key_size"`
	EC         string `json:"ec"`
	ECCurve    string `json:"ec_curve"`
}

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
