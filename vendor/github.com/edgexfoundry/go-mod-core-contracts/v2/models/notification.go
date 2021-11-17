//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

// Notification and its properties are defined in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/support-notifications/2.x#/Notification
// Model fields are same as the DTOs documented by this swagger. Exceptions, if any, are noted below.
type Notification struct {
	DBTimestamp
	Category    string
	Content     string
	ContentType string
	Description string
	Id          string
	Labels      []string
	Sender      string
	Severity    NotificationSeverity
	Status      NotificationStatus
}

// NotificationSeverity indicates the level of severity for the notification.
type NotificationSeverity string

// NotificationStatus indicates the current processing status of the notification.
type NotificationStatus string
