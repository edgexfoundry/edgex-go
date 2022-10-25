/*
 * Copyright (C) 2022 Canonical Ltd
 *
 *  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 * SPDX-License-Identifier: Apache-2.0'
 */

package main

import (
	"os"
)

func main() {
	// uncomment to enable snap debugging during development
	// snapctl.Set("debug", "true").Run()

	subCommand := os.Args[1]
	switch subCommand {
	case "install": // snap install hook
		install()
	case "configure": // snap configure hook
		configure()
	case "options": // apply snap options to apps
		options()
	default:
		panic("Unknown subcommand: " + subCommand)
	}
}
