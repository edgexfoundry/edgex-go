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
// It represents as a duration string. Will not do days you must compute to hours
//
// Nanosecond Duration = 1
// Microsecond = 1000 * Nanosecond
// Millisecond = 1000 * Microsecond
// Second = 1000 * Millisecond
// Minute = 60 * Second
// Hour = 60 * Minute
func parseFrequency(durationStr string) (time.Duration, error) {

	// Legacy ISO8601 format P1
	// ^P(\d+Y)?(\d+M)?(\d+D)?(T(\d+H)?(\d+M)?(\d+S)?)?$`
	matched, _ := regexp.MatchString(frequencyPattern, durationStr)
	if matched {
		if durationStr == "P" || durationStr == "PT" {
			matched = false
		}
		timeDuration := parseLegacyFrequency(durationStr)
		return timeDuration, nil
	}

	// GoLang time interval format
	// [-+]?([0-9]*(\.[0-9]*)?[a-z]+)+
	timeDuration, err := time.ParseDuration(durationStr)
	if err != nil {
		return 24 * time.Hour, err // default time.Duration w/error
	}

	return timeDuration, nil
}

func parseLegacyFrequency(durationStr string) time.Duration {
	durationRegex := regexp.MustCompile(`P(?P<years>\d+Y)?(?P<months>\d+M)?(?P<days>\d+D)?T?(?P<hours>\d+H)?(?P<minutes>\d+M)?(?P<seconds>\d+S)?`)
	matches := durationRegex.FindStringSubmatch(durationStr)

	years := parseInt64(matches[1])
	months := parseInt64(matches[2])
	days := parseInt64(matches[3])
	hours := parseInt64(matches[4])
	minutes := parseInt64(matches[5])
	seconds := parseInt64(matches[6])

	second := int64(time.Second)
	minute := int64(time.Minute)
	hour := int64(time.Hour)
	day := int64(24 * hour)
	month := int64(30 * day)
	year := int64(365 * day)
	return time.Duration(years*year + months*month + days*day + hours*hour + minutes*minute + seconds*second)
}

func parseInt64(value string) int64 {
	if len(value) == 0 {
		return 0
	}
	parsed, err := strconv.Atoi(value[:len(value)-1])
	if err != nil {
		return 0
	}
	return int64(parsed)
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
