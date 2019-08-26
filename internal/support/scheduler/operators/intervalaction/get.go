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
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type CollectionExecutor interface {
	Execute() ([]contract.IntervalAction, error)
}

type intervalActionLoadAll struct {
	database IntervalActionLoader
	limit    int
}

func (op intervalActionLoadAll) Execute() ([]contract.IntervalAction, error) {
	intervalActions, err := getIntervalActions(op.database, op.limit)
	if err != nil {
		return intervalActions, err
	}
	return intervalActions, nil
}

func NewAllExecutor(db IntervalActionLoader, limit int) CollectionExecutor {
	return intervalActionLoadAll{
		database: db,
		limit:    limit,
	}
}

func getIntervalActions(db IntervalActionLoader, limit int) ([]contract.IntervalAction, error) {
	var err error
	var intervalActions []contract.IntervalAction

	if limit <= 0 {
		intervalActions, err = db.IntervalActions()
	} else {
		intervalActions, err = db.IntervalActionsWithLimit(limit)
	}

	if err != nil {
		return nil, err
	}

	return intervalActions, err
}
