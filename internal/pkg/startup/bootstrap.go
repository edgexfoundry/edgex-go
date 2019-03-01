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
 *******************************************************************************/
package startup

import "sync"

type RetryFunc func(UseRegistry bool, useProfile string, timeout int, wait *sync.WaitGroup, ch chan error)

type LogFunc func(err error)

type BootParams struct {
	UseRegistry bool
	UseProfile  string
	BootTimeout int
}

func Bootstrap(params BootParams, retry RetryFunc, log LogFunc) {
	deps := make(chan error, 2)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go retry(params.UseRegistry, params.UseProfile, params.BootTimeout, &wg, deps)
	go func(ch chan error) {
		for {
			select {
			case e, ok := <-ch:
				if ok {
					log(e)
				} else {
					return
				}
			}
		}
	}(deps)

	wg.Wait()
}
