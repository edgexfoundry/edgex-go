/*******************************************************************************
 * Copyright 2023 Intel Corporation
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

package config

// BootStrapperInfo defines the first stage gate info
// It is the first stage gate of security bootstrapping
type BootStrapperInfo struct {
	Host      string
	StartPort int
}

// ReadyInfo defines the ready stage gate info
// It is the last stage gate of security bootstrapping for
// Kong, and all other Edgex core services
type ReadyInfo struct {
	ToRunPort int
}

// TokensInfo defines the tokens ready stage gate info
// for the secretstore setup (formerly known as vault-worker)
type TokensInfo struct {
	ReadyPort int
}

// SecretStoreSetupInfo defines the fields related to
// stage gating the secretstore setup (formerly known as vault-worker) bootstrapping
type SecretStoreSetupInfo struct {
	Host   string
	Tokens TokensInfo
}

// DatabaseInfo defines the fields related to
// stage gating the database bootstrapping
type DatabaseInfo struct {
	Host      string
	Port      int
	ReadyPort int
}

// RegistryInfo defines the fields related to
// stage gating the registry bootstrapping
type RegistryInfo struct {
	Host      string
	Port      int
	ReadyPort int
	ACL       ACLInfo
}

// ACLInfo defines the fields related to Registry's ACL process
type ACLInfo struct {
	// the protocol used for registry's API calls, usually it is different from the protocol of waitFor, i.e. TCP
	Protocol string
	// filepath to save the registry's token generated from ACL bootstrapping
	BootstrapTokenPath string
	// filepath for the secretstore's token created from secretstore-setup
	SecretsAdminTokenPath string
	// filepath for the sentinel file to indicate the registry ACL is set up successfully
	SentinelFilePath string
	// filepath to save the registry's token created for management purposes
	ManagementTokenPath string
	// the roles for registry role-based access control list
	Roles map[string]ACLRoleInfo
}

// ACLRoleInfo defines the fields related to Registry's ACL roles
type ACLRoleInfo struct {
	// the details about the role
	Description string
}

// KongDBInfo defines the fields related to
// stage gating the Kong's database bootstrapping
type KongDBInfo struct {
	Host      string
	Port      int
	ReadyPort int
}

// StageGateInfo defines the gate info for the security bootstrapper
// in different stages for services. From the YAML structure perspective,
// it is segmented in the way that environment variables are easier
// to read when they become all upper cases in the environment override.
type StageGateInfo struct {
	BootStrapper     BootStrapperInfo
	Ready            ReadyInfo
	SecretStoreSetup SecretStoreSetupInfo
	Database         DatabaseInfo
	Registry         RegistryInfo
	KongDB           KongDBInfo
	WaitFor          WaitForInfo
}
