/********************************************************************************
 *  Copyright (c) 2025 IOTech Ltd
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

package envelope

import (
	"encoding/json"
	"os"

	"github.com/fxamacker/cbor/v2"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

func Marshal(v any) ([]byte, error) {
	if os.Getenv(common.EnvMessageCborEncode) == common.ValueTrue {
		b, err := cbor.Marshal(v)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.Kind(err), "encode the message by CBOR", err)
		}
		return b, nil
	}

	b, err := json.Marshal(v)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(err), "encode the message by JSON", err)
	}
	return b, nil
}
