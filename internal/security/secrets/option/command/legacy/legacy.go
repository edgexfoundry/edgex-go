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

package legacy

import (
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option/command/constant"
)

type Command struct {
	configFile string
}

func NewCommand(configFile string) *Command {
	return &Command{
		configFile: configFile,
	}
}

func (c *Command) Execute() (statusCode int, err error) {
	err = option.GenTLSAssets(c.configFile)
	if err != nil {
		statusCode = constant.ExitWithError
	} else {
		statusCode = constant.ExitNormal
	}
	return
}
