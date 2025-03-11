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

package certtools

import (
	"crypto"
	"fmt"
	"github.com/openziti/identity/engines"
	"net/url"

	_ "github.com/openziti/identity/engines/parsec"
	_ "github.com/openziti/identity/engines/pkcs11"
)

func LoadEngineKey(engine string, addr *url.URL) (crypto.PrivateKey, error) {
	loadEngines()

	if eng, ok := engines.GetEngine(engine); ok {
		return eng.LoadKey(addr)
	} else {
		return nil, fmt.Errorf("engine '%s' is not supported", engine)
	}
}
