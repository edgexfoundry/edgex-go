//
// Copyright (c) 2022-2023 Intel Corporation
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
// SPDX-License-Identifier: Apache-2.0
//

package spiffetokenprovider

import (
	"context"

	"github.com/edgexfoundry/edgex-go/internal/security/common"
	securityCommon "github.com/edgexfoundry/edgex-go/internal/security/common"
	fileProviderConfig "github.com/edgexfoundry/edgex-go/internal/security/fileprovider/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-secrets/v3/secrets"
)

func makeToken(serviceName string,
	privilegedToken string,
	tokenConfig fileProviderConfig.TokenFileProviderInfo,
	secretStoreClient secrets.SecretStoreClient,
	lc logger.LoggingClient) (interface{}, error) {

	lc.Infof("generating policy/token defaults for service %s", serviceName)
	lc.Infof("using policy/token defaults for service %s", serviceName)
	servicePolicy := securityCommon.MakeDefaultTokenPolicy(serviceName)
	defaultPolicyPaths := servicePolicy["path"].(map[string]interface{})
	for pathKey, policy := range defaultPolicyPaths {
		servicePolicy["path"].(map[string]interface{})[pathKey] = policy
	}

	credentialGenerator := secretstore.NewDefaultCredentialGenerator()

	userManager := common.NewUserManager(lc, secretStoreClient, tokenConfig.UserPassMountPoint, "edgex-identity",
		privilegedToken, tokenConfig.DefaultTokenTTL, tokenConfig.DefaultJWTAudience, tokenConfig.DefaultJWTTTL)

	// Generate a random password

	randomPassword, err := credentialGenerator.Generate(context.TODO())
	if err != nil {
		return nil, err
	}

	// Create a user with the random password

	err = userManager.CreatePasswordUserWithPolicy(serviceName, randomPassword, "edgex-service-", servicePolicy)
	if err != nil {
		return nil, err
	}

	// Immediately log in the user to get a vault token

	var createTokenResponse interface{}
	if createTokenResponse, err = secretStoreClient.InternalServiceLogin(privilegedToken, tokenConfig.UserPassMountPoint, serviceName, randomPassword); err != nil {
		return nil, err
	}

	return createTokenResponse, nil
}
