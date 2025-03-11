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

import "sync/atomic"

type AtomicValue[T any] atomic.Value

func (self *AtomicValue[T]) Store(val T) {
	(*atomic.Value)(self).Store(val)
}

func (self *AtomicValue[T]) Load() T {
	var result T
	if val := (*atomic.Value)(self).Load(); val != nil {
		result = val.(T)
	}

	return result
}

func (self *AtomicValue[T]) CompareAndSwap(old, new T) bool {
	return (*atomic.Value)(self).CompareAndSwap(old, new)
}

func (self *AtomicValue[T]) Swap(new T) T {
	result := (*atomic.Value)(self).Swap(new)
	var old T
	if v, ok := result.(T); ok {
		old = v
	}
	return old
}
