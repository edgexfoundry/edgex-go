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

// StageGateInfo defines the gate info for the security bootstrapper
// in different stages for services. From the YAML structure perspective,
// it is segmented in the way that environment variables are easier
// to read when they become all upper cases in the environment override.
type StageGateInfo struct {
	BootStrapper     BootStrapperInfo
	Ready            ReadyInfo
	SecretStoreSetup SecretStoreSetupInfo
	Database         DatabaseInfo
	WaitFor          WaitForInfo
}
