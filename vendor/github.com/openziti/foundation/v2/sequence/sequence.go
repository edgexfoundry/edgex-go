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

package sequence

import (
	"fmt"
	"github.com/openziti/foundation/v2/info"
	"github.com/speps/go-hashids"
	"math/rand"
	"sync/atomic"
)

type Sequence struct {
	value  uint32
	hasher *hashids.HashIDData
}

func NewSequence() *Sequence {
	s := &Sequence{value: 0, hasher: hashids.NewData()}
	s.hasher.Salt = fmt.Sprintf("%d%d", info.NowInMilliseconds(), rand.Int63())
	s.hasher.MinLength = 4
	return s
}

func (s *Sequence) Next() uint32 {
	return atomic.AddUint32(&s.value, 1)
}

func (s *Sequence) NextHash() (string, error) {
	hasher, err := hashids.NewWithData(s.hasher)
	if err != nil {
		return "", err
	}
	next := s.Next()
	id, err := hasher.Encode([]int{int(next)})
	if err != nil {
		return "", err
	}
	return id, nil
}
