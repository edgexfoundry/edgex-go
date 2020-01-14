/*******************************************************************************
 * Copyright 2020 Dell Inc.
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

package acceptance

import (
	"context"
	"flag"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

const ENV_NO_SECURITY = "EDGEX_SECURITY_SECRET_STORE=false"

// MainFunc defines the signature of the function used to start a service.
type MainFunc func(ctx context.Context, cancel context.CancelFunc, router *mux.Router, readyStream chan<- bool)

// NewSUT implements the generic code to start a service within a separate goRoutine within the Golang test runner
// context.
func NewSUT(
	t *testing.T,
	env []string,
	args []string,
	mainFunc MainFunc) (context.CancelFunc, *sync.WaitGroup, *mux.Router) {

	setEnv := func(env []string) {
		os.Clearenv()
		for k := range env {
			kv := strings.Split(env[k], "=")
			_ = os.Setenv(kv[0], kv[1])
		}
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	router := mux.NewRouter()
	stream := make(chan bool, 1)

	wg.Add(1)
	go func() {
		origArgs := os.Args
		origEnv := os.Environ()

		os.Args = args
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		setEnv(env)

		mainFunc(ctx, cancel, router, stream)

		stream <- false
		close(stream)

		setEnv(origEnv)
		os.Args = origArgs
		wg.Done()
	}()

	result := <-stream
	if result == false {
		wg.Wait()
		assert.FailNow(t, "unable to start")
	}
	return cancel, &wg, router
}
