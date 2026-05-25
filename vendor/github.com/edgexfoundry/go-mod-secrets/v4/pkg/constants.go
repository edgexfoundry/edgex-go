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

package pkg

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
)

// Defines the valid secret store providers.
const (
	CoreSecurityServiceKey         = "edgex-core-security"
	ConfigFileName                 = "configuration.toml"
	ConfigDirectory                = "./res"
	ConfigDirEnv                   = "EDGEX_CONF_DIR"
	SpiffeTokenProviderGetTokenAPI = common.ApiBase + "/gettoken" // nolint: gosec
)
