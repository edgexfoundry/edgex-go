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
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/edgexfoundry/edgex-go/internal/system/executor"
)

func main() {
	output, err := executor.Execute(os.Args, func(args ...string) ([]byte, error) {
		return exec.Command("docker", args...).CombinedOutput()
	})
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	result, err := json.Marshal(output)
	if err != nil {
		fmt.Printf("json.Marshal error: %s", err.Error())
		os.Exit(1)
	}

	fmt.Print(string(result))
}
