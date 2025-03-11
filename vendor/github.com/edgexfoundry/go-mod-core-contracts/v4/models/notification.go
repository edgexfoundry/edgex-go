//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

type Notification struct {
	DBTimestamp
	Category     string
	Content      string
	ContentType  string
	Description  string
	Id           string
	Labels       []string
	Sender       string
	Severity     NotificationSeverity
	Status       NotificationStatus
	Acknowledged bool
}

// NotificationSeverity indicates the level of severity for the notification.
type NotificationSeverity string

// NotificationStatus indicates the current processing status of the notification.
type NotificationStatus string
