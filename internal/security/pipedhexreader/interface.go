//
// Copyright (c) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package pipedhexreader

// PipedHexReader is an interface to read hex bytes from
// standard output stream of an executable into a byte array
type PipedHexReader interface {
	// ReadHexBytesFromExe invokes executable
	// and reads hex bytes from stdout and returns an array
	ReadHexBytesFromExe(executablePath string) ([]byte, error)
}
