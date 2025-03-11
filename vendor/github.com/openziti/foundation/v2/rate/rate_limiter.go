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

// A RateLimiter allows running arbitrary, sequential operations with a limiter, so that only N operations
// can be queued to run at any given time. If the system is too busy, the rate limiter will return
// an ApiError indicating that the server is too busy
type RateLimiter interface {
	RunRateLimited(func() error) error
	GetQueueFillPct() float64
}

// An AdaptiveRateLimiter allows running arbitrary, sequential operations with a limiter, so that only N operations
// can be queued to run at any given time. If the system is too busy, the rate limiter will return
// an ApiError indicating that the server is too busy.
//
// The rate limiter returns a RateLimitControl, allow the calling code to indicate if the operation finished in
// time. If operations are timing out before the results are available, the rate limiter should allow fewer
// operations in, as they will likely time out before the results can be used.
//
// The rate limiter doesn't have a set queue size, it has a window which can grow and shrink. When
// a timeout is signaled, using the RateLimitControl, it shrinks the window based on queue position
// of the timed out operation. For example, if an operation was queued at position 200, but the times
// out, we assume that we need to limit the queue size to something less than 200 for now.
//
// The limiter will also reject already queued operations if the window size changes and the operation
// was queued at a position larger than the current window size.
//
// The window size will slowly grow back towards the max as successes are noted in the RateLimitControl.
type AdaptiveRateLimiter interface {
	RunRateLimited(f func() error) (RateLimitControl, error)
}

// An AdaptiveRateLimitTracker works similarly to an AdaptiveRateLimiter, except it just manages the rate
// limiting without actually running the work. Because it doesn't run the work itself, it has to account
// for the possibility that some work may never report as complete or failed. It thus has a configurable
// timeout at which point outstanding work will be marked as failed.
type AdaptiveRateLimitTracker interface {
	RunRateLimited(label string) (RateLimitControl, error)
	RunRateLimitedF(label string, f func(control RateLimitControl) error) error
	IsRateLimited() bool
}

type NoOpRateLimiter struct{}

func (self NoOpRateLimiter) RunRateLimited(f func() error) error {
	return f()
}

func (self NoOpRateLimiter) GetQueueFillPct() float64 {
	return 0
}

type NoOpAdaptiveRateLimiter struct{}

func (self NoOpAdaptiveRateLimiter) RunRateLimited(f func() error) (RateLimitControl, error) {
	return noOpRateLimitControl{}, f()
}

type NoOpAdaptiveRateLimitTracker struct{}

func (n NoOpAdaptiveRateLimitTracker) RunRateLimited(string) (RateLimitControl, error) {
	return noOpRateLimitControl{}, nil
}

func (n NoOpAdaptiveRateLimitTracker) RunRateLimitedF(_ string, f func(control RateLimitControl) error) error {
	return f(noOpRateLimitControl{})
}

func (n NoOpAdaptiveRateLimitTracker) IsRateLimited() bool {
	return false
}

type RateLimitControl interface {
	// Success indicates the operation was a success
	Success()

	// Backoff indicates that we need to backoff
	Backoff()

	// Failed indicates the operation was not a success, but a backoff isn't required
	Failed()
}

func NoOpRateLimitControl() RateLimitControl {
	return noOpRateLimitControl{}
}

type noOpRateLimitControl struct{}

func (noOpRateLimitControl) Success() {}

func (noOpRateLimitControl) Backoff() {}

func (noOpRateLimitControl) Failed() {}
