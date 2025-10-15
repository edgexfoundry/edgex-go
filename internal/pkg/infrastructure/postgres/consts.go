//
// Copyright (C) 2024-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

// constants relate to the postgres db schema names
const (
	coreDataSchema             = "core_data"
	coreKeeperSchema           = "core_keeper"
	coreMetaDataSchema         = "core_metadata"
	supportNotificationsSchema = "support_notifications"
	supportSchedulerSchema     = "support_scheduler"
	ProxyAuthSchema            = "security_proxy_auth"
)

// constants relate to the postgres db table names
const (
	configTableName               = coreKeeperSchema + ".config"
	eventTableName                = coreDataSchema + ".event"
	deviceInfoTableName           = coreDataSchema + ".device_info"
	deviceServiceTableName        = coreMetaDataSchema + ".device_service"
	deviceProfileTableName        = coreMetaDataSchema + ".device_profile"
	deviceTableName               = coreMetaDataSchema + ".device"
	provisionWatcherTableName     = coreMetaDataSchema + ".provision_watcher"
	notificationTableName         = supportNotificationsSchema + ".notification"
	readingTableName              = coreDataSchema + ".reading"
	registryTableName             = coreKeeperSchema + ".registry"
	scheduleActionRecordTableName = supportSchedulerSchema + ".record"
	scheduleJobTableName          = supportSchedulerSchema + ".job"
	subscriptionTableName         = supportNotificationsSchema + ".subscription"
	transmissionTableName         = supportNotificationsSchema + ".transmission"
	keyStoreTableName             = ProxyAuthSchema + ".key_store"
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
