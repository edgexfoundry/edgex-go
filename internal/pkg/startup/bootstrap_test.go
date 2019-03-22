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

import (
	"sync"
	"testing"

	"fmt"
	"time"
)

var checkInit bool
var checkLog string
var timeoutPass = 100
var timeoutFail = 1000
var wg sync.WaitGroup

func TestBootstrap(t *testing.T) {
	testPass(t)
	testFail(t)
}

func clearVars() {
	checkInit = false
	checkLog = ""
}

func testPass(t *testing.T) {
	clearVars()
	p := BootParams{true, "", timeoutPass}
	Bootstrap(p, mockRetry, mockLog)
	if !checkInit {
		t.Error("checkInit should be true.")
	}
	if checkLog != "" {
		t.Error("checkLog should be blank.")
	}
}

func testFail(t *testing.T) {
	clearVars()
	p := BootParams{true, "", timeoutFail}
	Bootstrap(p, mockRetry, mockLog)
	time.Sleep(time.Millisecond * time.Duration(25)) //goroutine timing
	if checkInit {
		t.Error("checkInit should be false.")
	}
	wg.Wait()
	if checkLog == "" {
		t.Error("checkLog should not be blank.")
	}
}

//Different test cases are toggled according to the timeout value
//SUCCESS = short duration 100ms
//FAIL = long duration 1000ms
func mockRetry(UseRegistry bool, useProfile string, timeout int, wait *sync.WaitGroup, ch chan error) {
	until := time.Now().Add(time.Millisecond * time.Duration(timeout))
	for time.Now().Before(until) {
		if timeout == timeoutFail {
			err := fmt.Errorf("Timeout Fail caught")
			ch <- err
			close(ch)
			wait.Done()
			return
		}
	}
	checkInit = true
	close(ch)
	wait.Done()

	return
}

func mockLog(err error) {
	wg.Add(1)
	checkLog = fmt.Sprintf("%v", err)
	wg.Done()
}
