/*
	Copyright NetFoundry Inc.

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

package ziti

import (
	"errors"
	"fmt"
	"strings"
)

type KeyAlgVar string

func (f *KeyAlgVar) String() string {
	return fmt.Sprint(string(*f))
}

func (f *KeyAlgVar) Set(value string) error {
	value = strings.ToUpper(value)
	if value != "EC" && value != "RSA" {
		return errors.New("invalid option -- must specify either 'EC' or 'RSA'")
	}
	*f = KeyAlgVar(value)
	return nil
}

func (f *KeyAlgVar) EC() bool {
	return f.Get() == "EC"
}

func (f *KeyAlgVar) RSA() bool {
	return f.Get() == "RSA"
}

func (f *KeyAlgVar) Get() string {
	return string(*f)
}

func (f *KeyAlgVar) Type() string {
	return "RSA|EC"
}
