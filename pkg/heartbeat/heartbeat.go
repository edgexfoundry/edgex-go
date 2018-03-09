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
	"time"

	"github.com/edgexfoundry/edgex-go/support/logging-client"
)

func Start(msg string, interval int, logger logger.LoggingClient) {
	go sendBeats(msg, interval, logger)
}

// Executes the basic heartbeat for all servicves. Writes entries to the supplied logger.
func sendBeats(heartbeatMsg string, interval int, logger logger.LoggingClient) {
	// Loop forever
	for true {
		logger.Info(heartbeatMsg)
		time.Sleep(time.Millisecond * time.Duration(interval)) // Sleep based on supplied interval
	}
}
