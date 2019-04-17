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
	IntervalKey     = db.Interval
	IntervalNameKey = db.Interval + ":name"
)

var intervalKeys = []string{IntervalKey, IntervalNameKey}

type Interval struct {
	contract.Interval
}

func NewInterval(from contract.Interval) (i Interval) {
	i.Interval = from
	return
}

func (i Interval) Add() (cmds []DbCommand) {
	cmds = make([]DbCommand, len(intervalKeys))
	for _, key := range intervalKeys {
		switch key {
		case IntervalKey:
			cmds = append(cmds, DbCommand{Command: "ZADD", Hash: key, Key: i.ID, Rank: i.Modified})
		case IntervalNameKey:
			cmds = append(cmds, DbCommand{Command: "HSET", Hash: key, Key: i.Name, Value: i.ID})
		}
	}
	return cmds
}

func (i Interval) Remove() (cmds []DbCommand) {
	cmds = make([]DbCommand, len(intervalKeys))
	for _, key := range intervalKeys {
		switch key {
		case IntervalKey:
			cmds = append(cmds, DbCommand{Command: "ZREM", Hash: key, Key: i.ID})
		case IntervalNameKey:
			cmds = append(cmds, DbCommand{Command: "HDEL", Key: i.Name})
		}
	}
	return cmds
}
