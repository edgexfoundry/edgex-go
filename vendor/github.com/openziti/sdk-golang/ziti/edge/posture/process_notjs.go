//go:build !js

/*
	Copyright 2019 NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package posture

import (
	"crypto/sha512"
	"fmt"
	"github.com/michaelquigley/pfxlog"
	"github.com/mitchellh/go-ps"
	"github.com/shirou/gopsutil/v3/process"
	"os"
	"path/filepath"
	"strings"
)

func Process(providedPath string) ProcessInfo {
	expectedPath := filepath.Clean(providedPath)

	processes, err := ps.Processes()

	if err != nil {
		pfxlog.Logger().Debugf("error getting Processes: %v", err)
	}

	if len(processes) == 0 {
		pfxlog.Logger().Warnf("total processes found was zero, this is unexpected")
	}

	for _, proc := range processes {
		if !isProcessPath(expectedPath, proc.Executable()) {
			continue
		}

		procDetails, err := process.NewProcess(int32(proc.Pid()))

		if err != nil {
			continue
		}

		executablePath, err := procDetails.Exe()

		if err != nil {
			continue
		}

		if strings.EqualFold(executablePath, expectedPath) {
			isRunning, _ := procDetails.IsRunning()
			file, err := os.ReadFile(executablePath)

			if err != nil {
				pfxlog.Logger().Warnf("could not read process executable file: %v", err)
				return ProcessInfo{
					IsRunning:          isRunning,
					Hash:               "",
					SignerFingerprints: nil,
				}
			}

			sum := sha512.Sum512(file)
			hash := fmt.Sprintf("%x", sum[:])

			signerFingerprints, err := getSignerFingerprints(executablePath)

			if err != nil {
				pfxlog.Logger().Warnf("could not read process signatures: %v", err)
				return ProcessInfo{
					IsRunning:          isRunning,
					Hash:               hash,
					SignerFingerprints: nil,
				}
			}

			return ProcessInfo{
				IsRunning:          isRunning,
				Hash:               hash,
				SignerFingerprints: signerFingerprints,
			}
		}
	}

	return ProcessInfo{
		IsRunning:          false,
		Hash:               "",
		SignerFingerprints: nil,
	}
}
