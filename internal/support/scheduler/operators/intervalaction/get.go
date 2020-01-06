/*******************************************************************************
 * Copyright 2019 VMware Inc.
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

package intervalaction

import (
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/errors"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

// This interface that returns a collection of interval actions
type IntervalActionsExecutor interface {
	Execute() ([]contract.IntervalAction, error)
}

// This type is for getting interval actions from the database with a given limit.
type intervalActionLoadAll struct {
	database IntervalActionLoader
	config   bootstrapConfig.ServiceInfo
}

// This method gets interval actions from the database.
func (op intervalActionLoadAll) Execute() ([]contract.IntervalAction, error) {
	intervalActions, err := op.database.IntervalActions()

	if err != nil {
		return intervalActions, err
	}
	if len(intervalActions) > op.config.MaxResultCount {
		return nil, errors.NewErrLimitExceeded(len(intervalActions))
	}
	return intervalActions, nil
}

// This factory method returns an executor used to get interval actions.
func NewAllExecutor(db IntervalActionLoader, config bootstrapConfig.ServiceInfo) IntervalActionsExecutor {
	return intervalActionLoadAll{
		database: db,
		config:   config,
	}
}
