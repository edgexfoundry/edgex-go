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
	"regexp"
	"strconv"
	"time"
)

const (
	frequencyPattern = `^P(\d+Y)?(\d+M)?(\d+D)?(T(\d+H)?(\d+M)?(\d+S)?)?$`
)

func isFrequencyValid(frequency string) bool {
	matched, _ := regexp.MatchString(frequencyPattern, frequency)
	if matched {
		if frequency == "P" || frequency == "PT" {
			matched = false
		}
	}
	return matched
}

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