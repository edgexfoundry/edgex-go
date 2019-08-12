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

package interval

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/errors"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type IdExecutor interface {
	Execute() (contract.Interval, error)
}

type CollectionExecutor interface {
	Execute() ([]contract.Interval, error)
}

type intervalLoadAll struct {
	database IntervalLoader
	limit    int
}

type intervalLoadById struct {
	database IntervalLoader
	id       string
}

type intervalLoadByName struct {
	database IntervalLoader
	name     string
}

func (op intervalLoadAll) Execute() ([]contract.Interval, error) {
	var err error
	var intervals []contract.Interval

	if op.limit <= 0 {
		intervals, err = op.database.Intervals()
	} else {
		intervals, err = op.database.IntervalsWithLimit(op.limit)
	}

	if err != nil {
		return nil, err
	}

	return intervals, err
}
func (op intervalLoadById) Execute() (contract.Interval, error) {
	res, err := op.database.IntervalById(op.id)
	if err != nil {
		if err == db.ErrNotFound {
			err = errors.NewErrIntervalNotFound(op.id)
		}
		return res, err
	}
	return res, nil
}

func (op intervalLoadByName) Execute() (contract.Interval, error) {
	res, err := op.database.IntervalByName(op.name)
	if err != nil {
		if err == db.ErrNotFound {
			err = errors.NewErrIntervalNotFound(op.name)
		}
		return res, err
	}
	return res, nil
}

func NewAllExecutor(db IntervalLoader, limit int) CollectionExecutor {
	return intervalLoadAll{
		database: db,
		limit:    limit,
	}
}

func NewIdExecutor(db IntervalLoader, id string) IdExecutor {
	return intervalLoadById{
		database: db,
		id:       id,
	}
}

func NewNameExecutor(db IntervalLoader, name string) IdExecutor {
	return intervalLoadByName{
		database: db,
		name:     name,
	}
}
