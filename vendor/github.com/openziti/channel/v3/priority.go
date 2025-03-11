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

package channel

type Priority uint32

const (
	Highest  = 0
	High     = 1024
	Standard = 4096
	Low      = 10240
)

type priorityHeap []Sendable

func (pq priorityHeap) Len() int {
	return len(pq)
}

// Less sorts by sequence if the priority is equivalent, otherwise sorts by priority. This ensures that we don't beat
// up the egress ordering buffer. We send packets in sequence order, unless priority is implicated.
func (pq priorityHeap) Less(i, j int) bool {
	if pq[i].Priority() == pq[j].Priority() {
		return pq[i].Sequence() < pq[j].Sequence()
	} else {
		return pq[i].Priority() < pq[j].Priority()
	}
}

func (pq priorityHeap) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *priorityHeap) Push(x interface{}) {
	pm := x.(Sendable)
	*pq = append(*pq, pm)
}

func (pq *priorityHeap) Pop() interface{} {
	old := *pq
	n := len(old)
	pm := old[n-1]
	*pq = old[0 : n-1]
	return pm
}
