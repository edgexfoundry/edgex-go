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

// DomainProvider supplies the Windows domain name that the device is joined to,
// used for domain membership posture checks.
type DomainProvider interface {
	GetDomain() string
}

// DomainFuncAsProvider converts a simple domain-returning function into a DomainProvider.
func DomainFuncAsProvider(f func() string) DomainProvider {
	return DomainProviderFunc(f)
}

// DomainProviderFunc is a function adapter that implements DomainProvider.
type DomainProviderFunc func() string

func (f DomainProviderFunc) GetDomain() string {
	return f()
}
