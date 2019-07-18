//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
// in compliance with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under
// the License.
//
// SPDX-License-Identifier: Apache-2.0'
//
package option

import "errors"

// OptionsExecutor is an executor to process the input options
type OptionsExecutor interface {
	ProcessOptions() (int, error)
	executeOptions(...func(*PkiInitOption) (exitCode, error)) (exitCode, error)
}

// PkiInitOption contains command line options for pki-init
type PkiInitOption struct {
	GenerateOpt bool
	ImportOpt   bool
	CacheOpt    bool
	CacheCAOpt  bool
	executor    OptionsExecutor
}

// NewPkiInitOption constructor to get options built for pki-init
func NewPkiInitOption(opts PkiInitOption) (ex OptionsExecutor, statusCode int, err error) {
	// import option cannot be attempted with other modes
	if opts.ImportOpt && opts.GenerateOpt {
		return ex, exitWithError.intValue(), errors.New("Cannot attempt import option with other modes")
	}
	// cache option cannot used with -generate or -import or -cacheca
	if opts.CacheOpt && (opts.GenerateOpt || opts.ImportOpt || opts.CacheCAOpt) {
		return ex, exitWithError.intValue(), errors.New("Cannot attempt cache option with other modes")
	}

	// cache CA option cannot used with -generate or -import or -cache
	if opts.CacheCAOpt && (opts.GenerateOpt || opts.ImportOpt || opts.CacheOpt) {
		return ex, exitWithError.intValue(), errors.New("Cannot attempt cache CA option with other modes")
	}
	opts.executor = &opts

	return &opts, normal.intValue(), nil
}

// ProcessOptions processes all the input options from the caller
func (pkiInitOpt *PkiInitOption) ProcessOptions() (int, error) {
	statusCode, err := pkiInitOpt.executor.executeOptions(
		Generate(),
		Import(),
		Cache(),
		CacheCA(),
	)

	return statusCode.intValue(), err
}

func (pkiInitOpt *PkiInitOption) executeOptions(executors ...func(*PkiInitOption) (exitCode, error)) (exitCode, error) {
	var statusCode exitCode
	var err error
	for _, executor := range executors {
		// stop execution as soon as any error occurs
		if statusCode, err = executor(pkiInitOpt); err != nil {
			break
		}
	}
	return statusCode, err
}
