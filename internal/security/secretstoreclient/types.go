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
 *
 *******************************************************************************/

package secretstoreclient

import (
	"fmt"
	"net/url"
)

type SecretServiceInfo struct {
	Scheme                      string
	Server                      string
	ServerName                  string
	Port                        int
	CertPath                    string
	CaFilePath                  string
	CertFilePath                string
	KeyFilePath                 string
	TokenFolderPath             string
	TokenFile                   string
	VaultSecretShares           int
	VaultSecretThreshold        int
	TokenProvider               string
	TokenProviderArgs           []string
	TokenProviderType           string
	TokenProviderAdminTokenPath string
	RevokeRootTokens            bool
}

func (s SecretServiceInfo) GetSecretSvcBaseURL() string {
	url := &url.URL{
		Scheme: s.Scheme,
		Host:   fmt.Sprintf("%s:%v", s.Server, s.Port),
		Path:   "/",
	}
	return url.String()
}
