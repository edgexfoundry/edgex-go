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

package genext

func EqualSlices[T comparable](a []T, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for idx, aVal := range a {
		if aVal != b[idx] {
			return false
		}
	}
	return true
}

func Contains[T comparable](a []T, val T) bool {
	_, found := FirstIndexOf(a, val)
	return found
}

func Remove[T comparable](a []T, val T) []T {
	var result []T
	for _, next := range a {
		if next != val {
			result = append(result, next)
		}
	}
	return result
}

func ContainsAll[T comparable](a []T, values ...T) bool {
	for _, val := range values {
		if !Contains(a, val) {
			return false
		}
	}
	return true
}

func ContainsAny[T comparable](a []T, values ...T) bool {
	for _, val := range values {
		if Contains(a, val) {
			return true
		}
	}
	return false
}

func FirstIndexOf[T comparable](a []T, val T) (int, bool) {
	for idx, s := range a {
		if s == val {
			return idx, true
		}
	}
	return -1, false
}

func Difference[T comparable](a, b []T) []T {
	mb := map[T]bool{}
	for _, x := range b {
		mb[x] = true
	}
	var ab []T
	for _, x := range a {
		if _, ok := mb[x]; !ok {
			ab = append(ab, x)
		}
	}
	return ab
}

func OrDefault[T any](val *T) T {
	if val == nil {
		return *new(T)
	}
	return *val
}

func SetToSlice[T comparable](values map[T]struct{}) []T {
	var result []T
	for val := range values {
		result = append(result, val)
	}
	return result
}

func SliceToSet[T comparable](values []T) map[T]struct{} {
	result := map[T]struct{}{}
	for _, val := range values {
		result[val] = struct{}{}
	}
	return result
}
