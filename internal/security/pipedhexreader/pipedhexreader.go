//
// Copyright (c) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package pipedhexreader

import (
	"bufio"
	"encoding/hex"
	"os/exec"
	"strings"
)

// pipedHexReader stores instance data for the pipedhexreader
type pipedHexReader struct{}

// NewPipedHexReader creates a new PipedHexReader
func NewPipedHexReader() PipedHexReader {
	return &pipedHexReader{}
}

// ReadHexBytesFromExe see interface.go
func (phr *pipedHexReader) ReadHexBytesFromExe(executablePath string) ([]byte, error) {
	sanitizedExecutable, err := exec.LookPath(executablePath)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(sanitizedExecutable)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	reader := bufio.NewReader(stdout)
	// We don't WANT a newline, but code defensively
	hexbytes, _ := reader.ReadString('\n')
	// Readstring returns non-nil error if delim is not present: ignore this
	// StdoutPipe usage is to Wait at the end of the reading logic
	// because it closes the readers automatically
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	hexbytes = strings.TrimSuffix(hexbytes, "\n")
	bytes, err := hex.DecodeString(hexbytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
