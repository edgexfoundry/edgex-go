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

package models

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

const (
	IntervalActionKey       = db.IntervalAction
	IntervalActionNameKey   = db.IntervalAction + ":name"
	IntervalActionParentKey = db.IntervalAction + ":parent"
	IntervalActionTargetKey = db.IntervalAction + ":target"
)

var intervalActionKeys = []string{IntervalActionKey, IntervalActionNameKey, IntervalActionParentKey, IntervalActionTargetKey}

type IntervalAction struct {
	contract.IntervalAction
}

func NewIntervalAction(from contract.IntervalAction) (ia IntervalAction) {
	ia.IntervalAction = from
	return
}

func (ia IntervalAction) Add() (cmds []DbCommand) {
	cmds = make([]DbCommand, len(intervalActionKeys))
	for _, key := range intervalActionKeys {
		switch key {
		case IntervalActionKey:
			cmds = append(cmds, DbCommand{Command: "ZADD", Hash: key, Key: ia.ID, Rank: ia.Modified})
		case IntervalActionNameKey:
			cmds = append(cmds, DbCommand{Command: "HSET", Hash: key, Key: ia.Name, Value: ia.ID})
		case IntervalActionParentKey:
			cmds = append(cmds, DbCommand{Command: "SADD", Hash: key + ":" + ia.Interval, Key: ia.ID})
		case IntervalActionTargetKey:
			cmds = append(cmds, DbCommand{Command: "SADD", Hash: key + ":" + ia.Target, Key: ia.ID})
		}
	}
	return cmds
}

func (ia IntervalAction) Remove() (cmds []DbCommand) {
	cmds = make([]DbCommand, len(intervalActionKeys))
	for _, key := range intervalKeys {
		switch key {
		case IntervalActionKey:
			cmds = append(cmds, DbCommand{Command: "ZREM", Hash: key, Key: ia.ID})
		case IntervalActionNameKey:
			cmds = append(cmds, DbCommand{Command: "HDEL", Hash: key, Key: ia.Name})
		case IntervalActionParentKey:
			cmds = append(cmds, DbCommand{Command: "SREM", Hash: key + ":" + ia.Interval, Key: ia.ID})
		case IntervalActionTargetKey:
			cmds = append(cmds, DbCommand{Command: "SREM", Hash: key + ":" + ia.Target, Key: ia.ID})
		}
	}
	return cmds
}
