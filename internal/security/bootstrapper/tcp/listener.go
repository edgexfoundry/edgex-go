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
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
)

const (
	connectionTimeout = 5 * time.Second
)

type TcpServer struct {
}

func NewTcpServer() *TcpServer {
	return &TcpServer{}
}

// StartListener instantiates a new listener on port and optional host if it is not empty
// returns error if failed to create a listener on the port number
func (tcpSrv *TcpServer) StartListener(port int, lc logger.LoggingClient, host string) error {
	lc.Debugf("Starting listener on port %d ...", port)

	trimmedHost := strings.TrimSpace(host)
	doneSrv := net.JoinHostPort(trimmedHost, strconv.Itoa(port))

	listener, err := net.Listen("tcp", doneSrv)
	if err != nil {
		// nolint: staticcheck
		return fmt.Errorf("Failed to create TCP listener: %v", err)
	}

	defer func() {
		_ = listener.Close()
	}()

	lc.Infof("Security bootstrapper starts listening on tcp://%s", doneSrv)
	for {
		conn, err := listener.Accept()
		if err != nil {
			lc.Errorf("found error when accepting connection: %v ! retry again in one second", err)
			time.Sleep(time.Second)
			continue
		}

		lc.Infof("Accepted connection on %s", doneSrv)

		// once reached here, the connection is established, and consider the semaphore on this port is raised
		go func(c *net.Conn) {
			defer func() {
				_ = (*c).Close()
			}()

			if err := handleConnection(*c); err != nil {
				lc.Warnf("failed to write through connection on %s: %v", doneSrv, err)
			}

			// intended process listener is done
			lc.Debugf("connection on port %d is done", port)
		}(&conn)
	}
}

func handleConnection(conn net.Conn) error {
	bufWriter := bufio.NewWriter(conn)
	datetime := time.Now().String()
	_ = conn.SetWriteDeadline(time.Now().Add(connectionTimeout))
	_, err := bufWriter.Write([]byte(datetime))

	return err
}
