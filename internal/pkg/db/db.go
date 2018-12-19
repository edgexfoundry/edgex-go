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
	"time"
)

const (
	// Databases
	MongoDB  = "mongodb"
	MemoryDB = "memorydb"

	// Data
	EventsCollection          = "event"
	ReadingsCollection        = "reading"
	ValueDescriptorCollection = "valueDescriptor"

	//Export
	ExportCollection = "exportConfiguration"

	//Logging
	LogsCollection = "logEntry"

	// Metadata
	Device           = "device"
	DeviceProfile    = "deviceProfile"
	DeviceService    = "deviceService"
	Addressable      = "addressable"
	Command          = "command"
	DeviceReport     = "deviceReport"
	ScheduleEvent    = "scheduleEvent"
	Schedule         = "schedule"
	ProvisionWatcher = "provisionWatcher"
)

var (
	ErrNotFound            = errors.New("Item not found")
	ErrUnsupportedDatabase = errors.New("Unsuppored database type")
	ErrInvalidObjectId     = errors.New("Invalid object ID")
	ErrNotUnique           = errors.New("Resource already exists")
	ErrCommandStillInUse   = errors.New("Command is still in use by device profiles")
	ErrSlugEmpty           = errors.New("Slug is nil or empty")
	ErrNameEmpty           = errors.New("Name is required")
)

type Configuration struct {
	DbType       string
	Host         string
	Port         int
	Timeout      int
	DatabaseName string
	Username     string
	Password     string
}

func MakeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
