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

// ProcessProvider supplies information about specific processes running on the device,
// including execution state, binary hash, and code signing details for process posture checks.
type ProcessProvider interface {
	GetProcessInfo(path string) ProcessInfo
}

// ProcessInfo contains details about a specific process including whether it's running,
// its binary hash, and code signing fingerprints.
type ProcessInfo struct {
	IsRunning          bool
	Hash               string
	SignerFingerprints []string
	QueryId            string
}

// ProcessInfoFunc is a function adapter that implements ProcessProvider.
type ProcessInfoFunc func(path string) ProcessInfo

func (f ProcessInfoFunc) GetProcessInfo(path string) ProcessInfo {
	return f(path)
}
