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

type AtomicBitSet uint32

func (self *AtomicBitSet) Set(index int, val bool) {
	done := false
	for !done {
		current := self.Load()
		next := setBitAtIndex(current, index, val)
		done = self.CompareAndSetAll(current, next)
	}
}

func (self *AtomicBitSet) IsSet(index int) bool {
	return isBitSetAtIndex(self.Load(), index)
}

func (self *AtomicBitSet) CompareAndSet(index int, current, next bool) bool {
	for {
		currentSet := self.Load()
		if isBitSetAtIndex(currentSet, index) != current {
			return false
		}
		nextSet := setBitAtIndex(currentSet, index, next)
		if self.CompareAndSetAll(currentSet, nextSet) {
			return true
		}
	}
}

func (self *AtomicBitSet) Store(val uint32) {
	atomic.StoreUint32((*uint32)(self), val)
}

func (self *AtomicBitSet) Load() uint32 {
	return atomic.LoadUint32((*uint32)(self))
}

func (self *AtomicBitSet) CompareAndSetAll(current, next uint32) bool {
	return atomic.CompareAndSwapUint32((*uint32)(self), current, next)
}

func setBitAtIndex(bitset uint32, index int, val bool) uint32 {
	if val {
		return bitset | (1 << index)
	}
	return bitset & ^(1 << index)
}

func isBitSetAtIndex(bitset uint32, index int) bool {
	return bitset&(1<<index) != 0
}
