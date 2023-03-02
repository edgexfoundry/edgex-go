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
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"flag"
	"os"

	"github.com/canonical/edgex-snap-hooks/v3/log"
	opt "github.com/canonical/edgex-snap-hooks/v3/options"
)

// options is called by the main function to configure options
func options() {
	flagset := flag.NewFlagSet("options", flag.ExitOnError)
	app := flagset.String("app", "", "Name of the app")
	flagset.Parse(os.Args[2:])

	log.SetComponentName("options")

	if *app == "" {
		log.Fatalf("Missing app name")
	}

	log.Info("Processing snap options for " + *app)
	if err := opt.ProcessAppCustomOptions(*app); err != nil {
		log.Fatalf("Could not process custom options: %v", err)
	}

}
