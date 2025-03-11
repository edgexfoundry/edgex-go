/*
	Copyright NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package rate

import (
	"fmt"
	"math/rand"
	"time"
)

func ExampleRateLimiter() {
	var limiter RateLimiter

	// code runs in the context of the rate limiter. There's a strict rate limit
	limitedFunc := func(limiter RateLimiter, a string) (int, error) {
		var result int
		err := limiter.RunRateLimited(func() error {
			result = len(a)
			return nil
		})
		return result, err
	}

	strlen, err := limitedFunc(limiter, "hello")
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
	} else {
		fmt.Printf("string length: %d\n", strlen)
	}
}

func ExampleAdaptiveRateLimiter() {
	var limiter AdaptiveRateLimiter

	// code runs in the context of the rate limiter, the rate limit adapts to successes and back-offs
	// when things aren't processed quickly enough
	limitedFunc := func(limiter AdaptiveRateLimiter, a string) (int, RateLimitControl, error) {
		var result int
		ctrl, err := limiter.RunRateLimited(func() error {
			time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
			result = len(a)
			return nil
		})
		return result, ctrl, err
	}

	start := time.Now()
	strlen, ctrl, err := limitedFunc(limiter, "hello")
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
	} else if time.Since(start) > 250*time.Millisecond {
		ctrl.Backoff()
		fmt.Printf("operation took too long\n")
	} else {
		ctrl.Success()
		fmt.Printf("string length: %d\n", strlen)
	}
}

func ExampleAdaptiveRateLimitTracker() {
	var limiter AdaptiveRateLimitTracker

	// code runs in its own context, the rate limit adapts to successes and back-offs
	// when things aren't processed quickly enough
	limitedFunc := func(limiter AdaptiveRateLimitTracker, a string) (int, RateLimitControl, error) {
		ctrl, err := limiter.RunRateLimited("strlen")
		time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
		result := len(a)
		return result, ctrl, err
	}

	start := time.Now()
	strlen, ctrl, err := limitedFunc(limiter, "hello")
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
	} else if time.Since(start) > 250*time.Millisecond {
		ctrl.Backoff()
		fmt.Printf("operation took too long\n")
	} else {
		ctrl.Success()
		fmt.Printf("string length: %d\n", strlen)
	}
}
