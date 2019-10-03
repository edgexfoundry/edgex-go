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
package scheduler

import (
	"strconv"
	"time"
)

// Convert millisecond string to Time
func msToTime(ms string) (time.Time, error) {
	msInt, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		// todo: support-scheduler will be removed later issue_650a
		t, err := time.Parse(TIMELAYOUT, ms)
		if err == nil {
			return t, nil
		}
		return time.Time{}, err
	}

	return time.Unix(0, msInt*int64(time.Millisecond)), nil
}

// Frequency indicates how often the specific resource needs to be polled.
// It represents as a duration string.
//
// Nanosecond Duration = 1
// Microsecond = 1000 * Nanosecond
// Millisecond = 1000 * Microsecond
// Second = 1000 * Millisecond
// Minute = 60 * Second
// Hour = 60 * Minute
func parseFrequency(durationStr string) (time.Duration, error) {

	// check Frequency
	timeDuration, err := time.ParseDuration(durationStr)
	if err != nil {
		return 24 * time.Hour, err // default time.Duration w/error
	}

	return timeDuration, nil
}

// Scheduler Queue Client
var currentQueueClient *QueueClient // Singleton used so that queueClient can use it to de-reference readings
type QueueClient struct {
}

// NewClient
func NewSchedulerQueueClient() *QueueClient {
	queueClient := &QueueClient{}
	currentQueueClient = queueClient // Set the singleton
	return queueClient
}
