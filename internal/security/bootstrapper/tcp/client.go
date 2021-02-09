/*******************************************************************************
 * Copyright 2021 Intel Corporation
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
 *******************************************************************************/

package tcp

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

const (
	dialTimeoutDuration = 5 * time.Second
)

// DialTcp will instantiate a new TCP dialer trying to connect to the TCP server specified by host and port
// host name can be empty as indicated in Golang's godoc: https://godoc.org/net#Dial
// port number must be greater than 0
func DialTcp(host string, port int, lc logger.LoggingClient) error {
	tcpHost := strings.TrimSpace(host)
	if port <= 0 {
		return fmt.Errorf("for tcp dial, port number must be greater than 0: %d", port)
	}

	tcpServerAddr := net.JoinHostPort(tcpHost, strconv.Itoa(port))

	for { // keep trying until server connects
		lc.Debugf("Trying to connecting to TCP server at address: %s", tcpServerAddr)
		c, err := net.DialTimeout("tcp", tcpServerAddr, dialTimeoutDuration)
		if err != nil {
			var opErr *net.OpError
			if errors.As(err, &opErr) && opErr.Op == "dial" {
				lc.Infof("TCP server %s may be not ready yet, retry in 1 second", tcpServerAddr)
				time.Sleep(time.Second)
				continue
			} else {
				return err
			}
		}
		defer func() {
			_ = c.Close()
		}()

		lc.Infof("Connected with TCP server %s", tcpServerAddr)

		// don't need to do anything else once it's connected in terms of response to the gating listener
		break
	}
	return nil
}
