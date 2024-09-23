//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

// constants relate to the postgres db schema names
const (
	coreDataSchema         = "core_data"
	coreKeeperSchema       = "core_keeper"
	coreMetaDataSchema     = "core_metadata"
	supportSchedulerSchema = "support_scheduler"
)

// constants relate to the postgres db table names
const (
	configTableName               = coreKeeperSchema + ".config"
	eventTableName                = coreDataSchema + ".event"
	deviceServiceTableName        = coreMetaDataSchema + ".device_service"
	deviceProfileTableName        = coreMetaDataSchema + ".device_profile"
	deviceTableName               = coreMetaDataSchema + ".device"
	provisionWatcherTableName     = coreMetaDataSchema + ".provision_watcher"
	readingTableName              = coreDataSchema + ".reading"
	registryTableName             = coreKeeperSchema + ".registry"
	scheduleActionRecordTableName = supportSchedulerSchema + ".record"
	scheduleJobTableName          = supportSchedulerSchema + ".job"
)

// constants relate to the common db table column names
const (
	contentCol  = "content"
	createdCol  = "created"
	idCol       = "id"
	modifiedCol = "modified"
	statusCol   = "status"
)

// constants relate to the event/reading postgres db table column names
const (
	deviceNameCol   = "devicename"
	resourceNameCol = "resourcename"
	profileNameCol  = "profilename"
	sourceNameCol   = "sourcename"
	originCol       = "origin"
	valueTypeCol    = "valuetype"
	unitsCol        = "units"
	tagsCol         = "tags"
	eventIdFKCol    = "event_id"
	valueCol        = "value"
	binaryValueCol  = "binaryvalue"
	mediaTypeCol    = "mediatype"
	objectValueCol  = "objectvalue"
)

// constants relate to the keeper postgres db table column names
const (
	keyCol = "key"
)

// constants relate to the schedule action record postgres db table column names
const (
	actionCol      = "action"
	actionIdCol    = "action_id"
	jobNameCol     = "job_name"
	scheduledAtCol = "scheduled_at"
)

// constants relate to the field names in the content column
const (
	labelsField    = "Labels"
	nameField      = "Name"
	serviceIdField = "ServiceId"
)
