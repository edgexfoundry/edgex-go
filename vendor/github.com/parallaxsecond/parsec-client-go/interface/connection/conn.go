// Copyright 2021 Contributors to the Parsec project.
// SPDX-License-Identifier: Apache-2.0

package connection

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
)

const defaultUnixSocketAddress = "unix:/run/parsec/parsec.sock"
const parsecEndpointEnvironmentVariable = "PARSEC_SERVICE_ENDPOINT"

// Connection represents a connection to the parsec service.
type Connection interface {
	// Open should be called before use
	Open() error
	// Methods to read, write and close the connection
	io.ReadWriteCloser
}

// type to manage unix socket connection
type unixConnection struct {
	rwc  io.ReadWriteCloser
	path string
}

// Read data from unix socket - conn must have been opened or an error will be returned
func (conn *unixConnection) Read(p []byte) (n int, err error) {
	if conn.rwc == nil {
		return 0, fmt.Errorf("reading closed connection")
	}
	return conn.rwc.Read(p)
}

// Write data to unix socket - conn must have been opened or an error will be returned
func (conn *unixConnection) Write(p []byte) (n int, err error) {
	if conn.rwc == nil {
		return 0, fmt.Errorf("writing closed connection")
	}
	return conn.rwc.Write(p)
}

// Close the unix socket
func (conn *unixConnection) Close() error {
	// We'll allow closing a closed connection
	if conn.rwc != nil {
		err := conn.rwc.Close()
		if err != nil {
			return err
		}
	}
	conn.rwc = nil
	return nil
}

// Opens the unix socket ready for read/write
func (conn *unixConnection) Open() error {
	rwc, err := net.Dial("unix", conn.path)

	if err != nil {
		return err
	}
	conn.rwc = rwc
	return nil
}

// NewDefaultConnection opens the default connection to the parsec service.
// This returns a Connection.  If the PARSEC_SERVICE_ENDPOINT environment
// variable is set, then this will be used to determine how to connect to the
// parsec service.  This must be a valid URL, and currently only urls of the form
// unix:/path are supported.
// if the PARSEC_SERVICE_ENDPOINT environment variable is not set, then the default of
// unix:/run/parsec/parsec.sock will be used
// Connection implementations are not guaranteed to be thread safe, so should not be used
// across threads.
func NewDefaultConnection() (Connection, error) {
	addressRawURL := os.Getenv(parsecEndpointEnvironmentVariable)
	if addressRawURL == "" {
		addressRawURL = defaultUnixSocketAddress
	}

	sockURL, err := url.Parse(addressRawURL)
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(sockURL.Scheme) {
	case "unix":
		return &unixConnection{
			path: sockURL.Path,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported url scheme %v", sockURL.Scheme)
	}
}
