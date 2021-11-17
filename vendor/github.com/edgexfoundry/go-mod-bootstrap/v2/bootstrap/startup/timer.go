/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright 2020 Intel Inc.
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

package startup

import (
	"time"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/environment"
)

// Timer contains references to dependencies required by the startup timer implementation.
type Timer struct {
	startTime time.Time
	duration  time.Duration
	interval  time.Duration
}

// NewStartUpTimer is a factory method that returns an initialized Timer receiver struct.
func NewStartUpTimer(serviceKey string) Timer {
	startup := environment.GetStartupInfo(serviceKey)

	return Timer{
		startTime: time.Now(),
		duration:  time.Second * time.Duration(startup.Duration),
		interval:  time.Second * time.Duration(startup.Interval),
	}
}

// NewTimer is a factory method that returns a Timer initialized with passed in duration and interval.
func NewTimer(duration int, interval int) Timer {
	return Timer{
		startTime: time.Now(),
		duration:  time.Second * time.Duration(duration),
		interval:  time.Second * time.Duration(interval),
	}
}

// SinceAsString returns the time since the timer was created as a string.
func (t Timer) SinceAsString() string {
	return time.Since(t.startTime).String()
}

// RemainingAsString returns the time remaining on the timer as a string.
func (t Timer) RemainingAsString() string {

	remaining := t.duration - time.Since(t.startTime)
	if remaining < 0 {
		remaining = 0
	}
	return remaining.String()
}

// HasNotElapsed returns whether or not the duration specified during construction has elapsed.
func (t Timer) HasNotElapsed() bool {
	return time.Now().Before(t.startTime.Add(t.duration))
}

// SleepForInterval pauses execution for the interval specified during construction.
func (t Timer) SleepForInterval() {
	time.Sleep(t.interval)
}
