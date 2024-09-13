//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

// constants relate to the postgres db schema names
const (
	coreDataSchema   = "core_data"
	coreKeeperSchema = "core_keeper"
)

// constants relate to the postgres db table names
const (
	eventTableName    = coreDataSchema + ".event"
	readingTableName  = coreDataSchema + ".reading"
	configTableName   = coreKeeperSchema + ".config"
	registryTableName = coreKeeperSchema + ".registry"
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
