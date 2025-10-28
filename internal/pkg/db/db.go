/*******************************************************************************
 * Copyright 2018 Dell Inc.
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

package db

import (
	"errors"
)

var (
	ErrNotFound            = errors.New("item not found")
	ErrUnsupportedDatabase = errors.New("unsupported database type")
	ErrInvalidObjectId     = errors.New("invalid object ID")
	ErrNotUnique           = errors.New("resource already exists")
	ErrCommandStillInUse   = errors.New("command is still in use by device profiles")
	ErrSlugEmpty           = errors.New("slug is nil or empty")
	ErrNameEmpty           = errors.New("name is required")
)

type Configuration struct {
	DbType          string
	Host            string
	Port            int
	Timeout         string
	DatabaseName    string
	Username        string
	Password        string
	BatchSize       int
	MaxConns        int32
	MaxConnIdleTime string
	MaxConnLifetime string
}
