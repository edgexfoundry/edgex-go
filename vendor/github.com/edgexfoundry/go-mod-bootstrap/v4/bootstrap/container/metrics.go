//
// Copyright (c) 2022 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package container

import (
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
)

// MetricsManagerInterfaceName contains the name of the metrics.Manager implementation in the DIC.
var MetricsManagerInterfaceName = di.TypeInstanceToName((*interfaces.MetricsManager)(nil))

// MetricsManagerFrom helper function queries the DIC and returns the metrics.Manager implementation.
func MetricsManagerFrom(get di.Get) interfaces.MetricsManager {
	manager, ok := get(MetricsManagerInterfaceName).(interfaces.MetricsManager)
	if !ok {
		return nil
	}

	return manager
}
