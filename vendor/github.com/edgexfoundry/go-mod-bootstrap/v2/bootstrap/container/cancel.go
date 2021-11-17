/*******************************************************************************
 * Copyright 2020 Intel Inc.
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

package container

import (
	"context"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
)

// CancelFuncName contains the name of the context.CancelFunc in the DIC.
var CancelFuncName = di.TypeInstanceToName((*context.CancelFunc)(nil))

// CancelFuncFrom helper function queries the DIC and returns the context.CancelFunc.
func CancelFuncFrom(get di.Get) context.CancelFunc {
	cancelFunc, ok := get(CancelFuncName).(context.CancelFunc)
	if !ok {
		return nil
	}

	return cancelFunc
}
