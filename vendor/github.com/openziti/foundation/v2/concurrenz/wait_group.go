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

package concurrenz

import (
	"reflect"
	"sync"
	"time"
)

// WaitGroup provides a facility to wait for an arbitrarily size collection of notification channels to be completed
//
//	The methods are multi-thread safe, but notifiers added after WaitForDone has been called are not guaranteed
//	to be waited for
type WaitGroup interface {
	// AddNotifier adds a notifier to the wait group
	AddNotifier(ch <-chan struct{})

	// WaitForDone will wait for all notifiers to complete up to the given deadline. It will return false if the timeout
	//             was reached, true otherwise
	WaitForDone(timeout time.Duration) bool
}

func NewWaitGroup() WaitGroup {
	return &waitGroupImpl{
		channels: map[reflect.Value]struct{}{},
	}
}

type waitGroupImpl struct {
	lock     sync.Mutex
	channels map[reflect.Value]struct{}
}

func (wg *waitGroupImpl) AddNotifier(ch <-chan struct{}) {
	wg.lock.Lock()
	defer wg.lock.Unlock()
	wg.channels[reflect.ValueOf(ch)] = struct{}{}
}

func (wg *waitGroupImpl) getSelectCases(timer <-chan time.Time) []reflect.SelectCase {
	wg.lock.Lock()
	defer wg.lock.Unlock()

	var cases []reflect.SelectCase
	cases = append(cases, reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(timer),
	})

	for ch := range wg.channels {
		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: ch,
		})
	}
	return cases
}

func (wg *waitGroupImpl) notiferComplete(v reflect.Value) {
	wg.lock.Lock()
	defer wg.lock.Unlock()
	delete(wg.channels, v)
}

func (wg *waitGroupImpl) WaitForDone(timeout time.Duration) bool {
	if len(wg.channels) == 0 {
		return true
	}

	timer := time.After(timeout)

	for len(wg.channels) > 0 {
		cases := wg.getSelectCases(timer)

		chosen, _, ok := reflect.Select(cases)
		if chosen == 0 {
			return false
		}
		if !ok {
			wg.notiferComplete(cases[chosen].Chan)
		}
	}
	return true
}
