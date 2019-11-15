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
	"flag"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/option/command/constant"
)

type FlagSet struct {
	flagSet    *flag.FlagSet
	configFile string
}

func NewFlags() (flags *FlagSet) {
	flags = &FlagSet{
		flagSet: flag.NewFlagSet(constant.CommandLegacy, flag.ExitOnError),
	}
	flags.flagSet.StringVar(&flags.configFile, "config", "", "specify JSON configuration file: /path/to/file.json")
	flags.flagSet.StringVar(&flags.configFile, "c", "", "specify JSON configuration file: /path/to/file.json")
	return
}
