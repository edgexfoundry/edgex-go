//
// Copyright (C) 2024-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	data "github.com/edgexfoundry/edgex-go/internal/core/data/embed"
	keeper "github.com/edgexfoundry/edgex-go/internal/core/keeper/embed"
	metadata "github.com/edgexfoundry/edgex-go/internal/core/metadata/embed"
	proxyauth "github.com/edgexfoundry/edgex-go/internal/security/proxyauth/embed"
	notifications "github.com/edgexfoundry/edgex-go/internal/support/notifications/embed"
	scheduler "github.com/edgexfoundry/edgex-go/internal/support/scheduler/embed"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
)

// constants relate to the postgres db table names
const (
	configTableName               = keeper.SchemaName + ".config"
	eventTableName                = data.SchemaName + ".event"
	deviceInfoTableName           = data.SchemaName + ".device_info"
	deviceServiceTableName        = metadata.SchemaName + ".device_service"
	deviceProfileTableName        = metadata.SchemaName + ".device_profile"
	deviceTableName               = metadata.SchemaName + ".device"
	provisionWatcherTableName     = metadata.SchemaName + ".provision_watcher"
	notificationTableName         = notifications.SchemaName + ".notification"
	readingTableName              = data.SchemaName + ".reading"
	registryTableName             = keeper.SchemaName + ".registry"
	scheduleActionRecordTableName = scheduler.SchemaName + ".record"
	scheduleJobTableName          = scheduler.SchemaName + ".job"
	subscriptionTableName         = notifications.SchemaName + ".subscription"
	transmissionTableName         = notifications.SchemaName + ".transmission"
	keyStoreTableName             = proxyauth.SchemaName + ".key_store"
)

// constants relate to the common db table column names
const (
	contentCol  = "content"
	createdCol  = "created"
	idCol       = "id"
	modifiedCol = "modified"
	statusCol   = "status"
	nameCol     = "name"
)

// constants relate to the named arguments as specified for SQL conditions
const (
	offsetCondition      = common.Offset
	limitCondition       = common.Limit
	startTimeCondition   = common.Start
	endTimeCondition     = common.End
	jsonContentCondition = "jsonContent"
	categoryCondition    = "category"
	labelsCondition      = "labels"
)

// constants relate to the event/reading postgres db table column names
const (
	deviceNameCol     = "devicename"
	resourceNameCol   = "resourcename"
	profileNameCol    = "profilename"
	sourceNameCol     = "sourcename"
	originCol         = "origin"
	valueTypeCol      = "valuetype"
	unitsCol          = "units"
	tagsCol           = "tags"
	eventIdFKCol      = "event_id"
	deviceInfoIdFKCol = "device_info_id"
	valueCol          = "value"
	numericValueCol   = "numeric_value"
	binaryValueCol    = "binaryvalue"
	mediaTypeCol      = "mediatype"
	objectValueCol    = "objectvalue"
	markDeletedCol    = "mark_deleted"
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

// constants relate to the notification postgres db table column names
const (
	notificationIdCol = "notification_id"
)

// constants relate to the field names in the content column
const (
	categoryField         = "Category"
	categoriesField       = "Categories"
	createdField          = "Created"
	labelsField           = "Labels"
	parentField           = "Parent"
	manufacturerField     = "Manufacturer"
	modelField            = "Model"
	nameField             = "Name"
	notificationIdField   = "NotificationId"
	profileNameField      = "ProfileName"
	receiverField         = "Receiver"
	serviceIdField        = "ServiceId"
	serviceNameField      = "ServiceName"
	statusField           = "Status"
	subscriptionNameField = "SubscriptionName"
	acknowledgedField     = "Acknowledged"
)
