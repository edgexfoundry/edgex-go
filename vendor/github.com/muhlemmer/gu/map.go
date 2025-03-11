package gu

// MapEqual check if two maps have exactly the same content.
// If both maps are nil, they are considered equal.
// When a nil map is compared to an empty map,
// they are not considered equal.
func MapEqual[K, V comparable](a, b map[K]V) bool {
	if a == nil && b == nil {
		return true
	}

	if (a == nil && b != nil) || (b == nil && a != nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for k, av := range a {
		if bv, ok := b[k]; !ok || av != bv {
			return false
		}
	}

	return true
}

// MapCopy copies all the entries of src into a new map.
// Nil is returned when src is nil.
// Note that if V is a pointer or reference type,
// it is only shallow copied.
func MapCopy[K comparable, V any](src map[K]V) map[K]V {
	if src == nil {
		return nil
	}

	dst := make(map[K]V, len(src))

	for k, v := range src {
		dst[k] = v
	}

	return dst
}

// MapCopyKeys copies the entries or src,
// identified by keys into a new map.
// Nil is returned when src is nil.
//
// If no keys are provided and src is not nil,
// an empty non-nil map is returned.
//
// Note that if V is a pointer or reference type,
// it is only shallow copied.
func MapCopyKeys[K comparable, V any](src map[K]V, keys ...K) map[K]V {
	if src == nil {
		return nil
	}

	dst := make(map[K]V, len(keys))

	for _, k := range keys {
		if v, ok := src[k]; ok {
			dst[k] = v
		}
	}

	return dst
}

// MapMerge copies all entries from src in dst.
// Any pre-existing keys in dst are overwritten.
func MapMerge[K comparable, V any](src map[K]V, dst map[K]V) {
	for k, v := range src {
		dst[k] = v
	}
}
