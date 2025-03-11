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

import "sync"

type CopyOnWriteMap[K comparable, V any] struct {
	value AtomicValue[map[K]V]
	lock  sync.Mutex
}

func (self *CopyOnWriteMap[K, V]) Put(key K, value V) {
	self.lock.Lock()
	defer self.lock.Unlock()

	var current = self.value.Load()
	mapCopy := map[K]V{}
	for k, v := range current {
		mapCopy[k] = v
	}
	mapCopy[key] = value
	self.value.Store(mapCopy)
}

func (self *CopyOnWriteMap[K, V]) Get(key K) V {
	return self.value.Load()[key]
}

func (self *CopyOnWriteMap[K, V]) Delete(key K) {
	self.lock.Lock()
	defer self.lock.Unlock()

	var current = self.value.Load()
	mapCopy := map[K]V{}
	for k, v := range current {
		if k != key {
			mapCopy[k] = v
		}
	}
	self.value.Store(mapCopy)
}

func (self *CopyOnWriteMap[K, V]) AsMap() map[K]V {
	return self.value.Load()
}

func (self *CopyOnWriteMap[K, V]) Clear() {
	self.lock.Lock()
	defer self.lock.Unlock()
	self.value.Store(map[K]V{})
}

func (self *CopyOnWriteMap[K, V]) DeleteIf(f func(key K, val V) bool) bool {
	self.lock.Lock()
	defer self.lock.Unlock()

	matched := false
	var current = self.value.Load()
	mapCopy := map[K]V{}
	for k, v := range current {
		if !f(k, v) {
			mapCopy[k] = v
		} else {
			matched = true
		}
	}
	self.value.Store(mapCopy)
	return matched
}
