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

package xgress

import (
	"fmt"
	"github.com/emirpasic/gods/trees/btree"
	"github.com/emirpasic/gods/utils"
	"github.com/michaelquigley/pfxlog"
	"sync"
	"sync/atomic"
)

type LinkReceiveBuffer struct {
	sync.Mutex
	tree        *btree.Tree
	sequence    int32
	maxSequence int32
	size        uint32
	txQueue     chan *Payload
}

func NewLinkReceiveBuffer(txQueueSize int32) *LinkReceiveBuffer {
	return &LinkReceiveBuffer{
		tree:     btree.NewWith(10240, utils.Int32Comparator),
		sequence: -1,
		txQueue:  make(chan *Payload, txQueueSize),
	}
}

func (buffer *LinkReceiveBuffer) Size() uint32 {
	return atomic.LoadUint32(&buffer.size)
}

func (buffer *LinkReceiveBuffer) ReceiveUnordered(x *Xgress, payload *Payload, maxSize uint32) bool {
	buffer.Lock()
	defer buffer.Unlock()

	if payload.GetSequence() <= buffer.sequence {
		x.dataPlane.GetMetrics().MarkDuplicatePayload()
		return true
	}

	if atomic.LoadUint32(&buffer.size) > maxSize && payload.Sequence > buffer.maxSequence {
		x.dataPlane.GetMetrics().MarkPayloadDropped()
		return false
	}

	treeSize := buffer.tree.Size()
	buffer.tree.Put(payload.GetSequence(), payload)
	if buffer.tree.Size() > treeSize {
		payloadSize := len(payload.Data)
		size := atomic.AddUint32(&buffer.size, uint32(payloadSize))
		pfxlog.Logger().Tracef("Payload %v of size %v added to transmit buffer. New size: %v", payload.Sequence, payloadSize, size)
		if payload.Sequence > buffer.maxSequence {
			buffer.maxSequence = payload.Sequence
		}
	} else {
		x.dataPlane.GetMetrics().MarkDuplicatePayload()
	}

	buffer.queueNext()

	return true
}

func (buffer *LinkReceiveBuffer) queueNext() {
	if val := buffer.tree.LeftValue(); val != nil {
		payload := val.(*Payload)
		if payload.Sequence == buffer.sequence+1 {
			select {
			case buffer.txQueue <- payload:
				buffer.tree.Remove(payload.Sequence)
				buffer.sequence = payload.Sequence
			default:
			}
		}
	}
}

func (buffer *LinkReceiveBuffer) NextPayload(closeNotify <-chan struct{}) *Payload {
	select {
	case payload := <-buffer.txQueue:
		return payload
	default:
	}

	buffer.Lock()
	buffer.queueNext()
	buffer.Unlock()

	select {
	case payload := <-buffer.txQueue:
		return payload
	case <-closeNotify:
	}

	// closed, check if there's anything pending in the queue
	select {
	case payload := <-buffer.txQueue:
		return payload
	default:
		return nil
	}
}

func (buffer *LinkReceiveBuffer) Inspect() *RecvBufferDetail {
	buffer.Lock()
	defer buffer.Unlock()

	nextPayload := "none"
	if head := buffer.tree.LeftValue(); head != nil {
		payload := head.(*Payload)
		nextPayload = fmt.Sprintf("%v", payload.Sequence)
	}

	return &RecvBufferDetail{
		Size:           buffer.Size(),
		PayloadCount:   uint32(buffer.tree.Size()),
		Sequence:       buffer.sequence,
		MaxSequence:    buffer.maxSequence,
		NextPayload:    nextPayload,
		AcquiredSafely: true,
	}
}
