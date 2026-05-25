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
)

type CopyOnWriteSlice[T any] struct {
	value AtomicValue[[]T]
	lock  sync.Mutex
}

func (self *CopyOnWriteSlice[T]) Value() []T {
	return self.value.Load()
}

func (self *CopyOnWriteSlice[T]) Append(toAdd T) {
	self.lock.Lock()
	defer self.lock.Unlock()

	currentSlice := self.value.Load()
	newSlice := append(currentSlice, toAdd)
	self.value.Store(newSlice)
}

func (self *CopyOnWriteSlice[T]) Delete(toRemove T) {
	self.lock.Lock()
	defer self.lock.Unlock()

	currentSlice := self.value.Load()
	newSlice := make([]T, 0, len(currentSlice))
	for _, val := range currentSlice {
		if reflect.ValueOf(val).Interface() != reflect.ValueOf(toRemove).Interface() {
			newSlice = append(newSlice, val)
		}
	}
	self.value.Store(newSlice)
}

func (self *CopyOnWriteSlice[T]) DeleteIf(filter func(T) bool) {
	self.lock.Lock()
	defer self.lock.Unlock()

	currentSlice := self.value.Load()
	newSlice := make([]T, 0, len(currentSlice))
	for _, val := range currentSlice {
		if !filter(val) {
			newSlice = append(newSlice, val)
		}
	}
	self.value.Store(newSlice)
}
