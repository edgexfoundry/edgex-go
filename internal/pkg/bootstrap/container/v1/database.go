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

package v1

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/interfaces"

	"github.com/edgexfoundry/go-mod-bootstrap/di"
)

// DBClientInterfaceName contains the name of the interfaces.DBClient implementation in the DIC.
var DBClientInterfaceName = di.TypeInstanceToName((*interfaces.DBClient)(nil))

// DBClientFrom helper function queries the DIC and returns the interfaces.DBClient implementation.
func DBClientFrom(get di.Get) interfaces.DBClient {
	return get(DBClientInterfaceName).(interfaces.DBClient)
}
