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

import (
	"fmt"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/security/setup"
	"github.com/fsnotify/fsnotify"
)

// Import deploys PKI from CacheDir to DeployDir.  If CacheDir is empty,
// import blocks until the PKI assets shown in CacheDir;
// otherwise it just copies PKI assets from CacheDir into DeployDir.
func Import() func(*PkiInitOption) (exitCode, error) {
	return func(pkiInitOpton *PkiInitOption) (exitCode, error) {

		if isImportNoOp(pkiInitOpton) {
			return normal, nil
		}

		return importPkis()
	}
}

func isImportNoOp(pkiInitOption *PkiInitOption) bool {
	// nop: if the flag is missing or not on
	return pkiInitOption == nil || !pkiInitOption.ImportOpt
}

func importPkis() (statusCode exitCode, err error) {
	pkiCacheDir := getPkiCacheDirEnv()
	setup.LoggingClient.Info(fmt.Sprintf("PKI_CACHE: %s", pkiCacheDir))

	dirEmpty, err := isDirEmpty(pkiCacheDir)

	if err != nil {
		return exitWithError, err
	}

	if dirEmpty {

		pkiCacheWatcher, err := fsnotify.NewWatcher()

		if err != nil {
			return exitWithError, err
		}
		defer pkiCacheWatcher.Close()

		done := make(chan bool)

		go func() {
			for {
				select {
				case event := <-pkiCacheWatcher.Events:
					if event.Name == "" { // skip if event name is empty
						continue
					}
					setup.LoggingClient.Debug(fmt.Sprintf("watcher event: %#v\n", event))
					// wait for some time before the directory settle down on the source side
					// as the whole directory tree is just dropped in
					time.Sleep(3 * time.Second)

					err = deploy(pkiCacheDir, pkiInitDeployDir)
					if err != nil {
						statusCode = exitWithError
					} else {
						statusCode = normal
					}
					done <- true
				case watcherErr := <-pkiCacheWatcher.Errors:
					if watcherErr == nil {
						continue
					}
					setup.LoggingClient.Error(fmt.Sprintf("watcher error: %v\n", watcherErr))
					statusCode = exitWithError
					err = watcherErr
					done <- true
				}
			}
		}()

		if err := pkiCacheWatcher.Add(pkiCacheDir); err != nil {
			return exitWithError, err
		}

		<-done
	} else {
		// copy stuff into dest dir from pkiCache
		err = deploy(pkiCacheDir, pkiInitDeployDir)
		if err != nil {
			statusCode = exitWithError
		}
	}

	return statusCode, err
}
