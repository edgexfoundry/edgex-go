/*******************************************************************************
 * Copyright 2020 Dell Inc.
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

package test

import "time"

// Timer contains references to dependencies required by the acceptance timer implementation.
type Timer struct {
	startTime time.Time
	endTime   *time.Time
}

// NewTimer is a factory function that starts a timer and returns an initialized Timer receiver struct.
func NewTimer() *Timer {
	return &Timer{
		startTime: time.Now(),
		endTime:   nil,
	}
}

// Stop ends the timer.
func (t *Timer) Stop() {
	var now = time.Now()
	t.endTime = &now
}

// insideDeviation returns the time since NewTimer factory function was called in milliseconds.
func (t *Timer) insideDeviation(expectedInMS, deviationInMS int) bool {
	if t.endTime == nil {
		t.Stop()
	}

	init := func(value int) time.Duration {
		if value < 0 {
			return time.Duration(0)
		}
		return time.Duration(value)
	}
	elapsed := (*t.endTime).Sub(t.startTime) / time.Millisecond

	// verify elapsed is within the specified expectedInMS value (plus or minus).
	return init(expectedInMS-deviationInMS) <= elapsed && elapsed < init(expectedInMS+deviationInMS)
}
