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
 //TODO: Follow up with Jim and the community on the discussion of whether or not this actually belongs here.
 //      I believe we need some kind of common package for basic service capabilities. "Support" has a
 //      particular connotation as Jim describes it -- optional services that can be replaced by a 3rd party.
 //      This is not that.

package heartbeat

import (
	"time"

	"github.com/edgexfoundry/edgex-go/support/logging-client"
)

func Start(msg string, interval int, logger logger.LoggingClient) {
	chBeats := make(chan string)
	go sendBeats(msg, interval, chBeats)
	go func() {
		for {
			msg, ok := <-chBeats
			if !ok {
				break
			}
			logger.Info(msg)
		}
		close(chBeats)
	}()
}

// Executes the basic heartbeat for all servicves. Writes entries to the supplied channel.
func sendBeats(heartbeatMsg string, interval int, beats chan<- string) {
	// Loop forever
	for true {
		beats <- heartbeatMsg
		time.Sleep(time.Millisecond * time.Duration(interval)) // Sleep based on supplied interval
	}
}