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
 *
 * @author: Trevor Conn, Dell
 * @version: 0.5.0
 *******************************************************************************/

package heartbeat

import (
	"testing"
	"time"
)

type HeartbeatLogger struct {

}

var beatCount int

func TestHeartbeat(t *testing.T) {

	go Start("This is a test", 500, HeartbeatLogger{})
	stop := time.Now().Add(time.Millisecond * time.Duration(2000))
	for ; time.Now().Before(stop); time.Sleep(time.Millisecond * time.Duration(100)) {
		if beatCount > 0 {
			break
		}
	}

	if beatCount == 0 {
		t.Error("No heartbeat received. Waited 2 sec.")
	}
}

// Log an INFO level message
func (lc HeartbeatLogger) Info(msg string, labels ...string) error {
	beatCount++
	return nil
}

// Log a TRACE level message
func (lc HeartbeatLogger) Trace(msg string, labels ...string) error {
	return nil
}

// Log a DEBUG level message
func (lc HeartbeatLogger) Debug(msg string, labels ...string) error {
	return nil
}

// Log a WARN level message
func (lc HeartbeatLogger) Warn(msg string, labels ...string) error {
	return nil
}

// Log an ERROR level message
func (lc HeartbeatLogger) Error(msg string, labels ...string) error {
	return nil
}