// Copyright 2026 AxonOps Limited.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package syncmap

import "sync"

// SyncMap is a type-safe, generic wrapper around [sync.Map].
//
// The zero value is an empty map ready for use. SyncMap must not be
// copied after first use.
type SyncMap[K comparable, V any] struct {
	syncMap sync.Map
}

// Load returns the value stored in the map for key, or the zero value
// of V if no entry is present. The ok result reports whether an entry
// was found.
func (m *SyncMap[K, V]) Load(key K) (value V, ok bool) {
	v, ok := m.syncMap.Load(key)
	if !ok {
		var v2 V
		return v2, false
	}
	return v.(V), ok
}

// Store sets the value associated with key.
func (m *SyncMap[K, V]) Store(key K, value V) {
	m.syncMap.Store(key, value)
}

// LoadOrStore returns the existing value for key if present.
// Otherwise it stores value and returns it.
// The loaded result is true if the value was loaded, false if stored.
func (m *SyncMap[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	a, l := m.syncMap.LoadOrStore(key, value)
	return a.(V), l
}

// LoadAndDelete deletes the entry for key and returns its previous
// value, if any. The loaded result reports whether the key was
// present. If the key was not present, value is the zero value of V.
func (m *SyncMap[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	v, l := m.syncMap.LoadAndDelete(key)
	if !l {
		var v2 V
		return v2, false
	}
	return v.(V), l
}

// Delete removes the entry for key. It is a no-op if the key is not
// present.
func (m *SyncMap[K, V]) Delete(key K) {
	m.syncMap.Delete(key)
}

// Swap replaces the value stored for key with value and returns the
// previous value, if any. The loaded result reports whether the key
// was present. If the key was not present, previous is the zero
// value of V.
func (m *SyncMap[K, V]) Swap(key K, value V) (previous V, loaded bool) {
	v, l := m.syncMap.Swap(key, value)
	if !l {
		var v2 V
		return v2, false
	}
	return v.(V), l
}

// Clear removes all entries from the map, leaving it empty.
func (m *SyncMap[K, V]) Clear() {
	m.syncMap.Clear()
}

// CompareAndSwap swaps the old and new values for key if the value
// currently stored in m is equal to old. The swapped result reports
// whether the swap was performed.
//
// V must be comparable. Because SyncMap is declared with V any to
// support non-comparable value types, this operation cannot be a
// method on SyncMap[K, V]; instantiating it with a non-comparable V
// (slice, map, func, or a struct containing one of those) produces a
// compile-time error rather than the runtime panic that the
// underlying [sync.Map.CompareAndSwap] would raise.
//
// If V is itself an interface type, the comparison performed inside
// [sync.Map] can still panic at runtime when either operand's dynamic
// type is not comparable. This matches Go's `==` semantics for
// interfaces and is outside this wrapper's control.
func CompareAndSwap[K, V comparable](m *SyncMap[K, V], key K, old, new V) (swapped bool) {
	return m.syncMap.CompareAndSwap(key, old, new)
}

// CompareAndDelete deletes the entry for key if its current value is
// equal to old. The deleted result reports whether the entry was
// removed.
//
// V must be comparable, for the same reason as [CompareAndSwap].
func CompareAndDelete[K, V comparable](m *SyncMap[K, V], key K, old V) (deleted bool) {
	return m.syncMap.CompareAndDelete(key, old)
}

// Range calls f sequentially for each key and value present in the
// map. If f returns false, Range stops iteration.
//
// Range does not correspond to a consistent snapshot of the map's
// contents: no key will be visited more than once, but if a value is
// stored or deleted concurrently (including by f), Range may reflect
// any mapping for that key during the iteration.
//
// Range may run in O(n) time even if f returns false after a constant
// number of calls, where n is the number of elements in the map at
// the start of the call.
func (m *SyncMap[K, V]) Range(f func(key K, value V) bool) {
	m.syncMap.Range(func(key, value any) bool {
		return f(key.(K), value.(V))
	})
}

// Len returns the number of entries in the map at the moment of the
// call. It runs in O(n) time by traversing the map with Range.
//
// Because the traversal is not atomic, concurrent stores and deletes
// may cause the returned count to differ from the number of entries
// visible to any single subsequent operation. Treat the result as an
// approximation, not a consistent snapshot.
func (m *SyncMap[K, V]) Len() int {
	l := 0
	m.syncMap.Range(func(key, value any) bool {
		l++
		return true
	})
	return l
}

// Map returns a shallow copy of the map's contents as a plain Go map.
// It runs in O(n) time.
//
// The returned map is a point-in-time approximation: because the
// underlying Range traversal is not atomic, concurrent modifications
// may or may not be reflected in the result. The caller owns the
// returned map and may modify it freely.
func (m *SyncMap[K, V]) Map() map[K]V {
	newMap := make(map[K]V)
	m.Range(func(key K, value V) bool {
		newMap[key] = value
		return true
	})
	return newMap
}

// Keys returns a slice of all keys present in the map at the moment
// of the call. It runs in O(n) time.
//
// The result is a point-in-time approximation. Concurrent stores and
// deletes may cause the slice to include keys that have since been
// removed, or to omit keys that were added during traversal. The
// order of keys is undefined.
func (m *SyncMap[K, V]) Keys() []K {
	var keys []K
	m.syncMap.Range(func(key, value any) bool {
		keys = append(keys, key.(K))
		return true
	})
	return keys
}

// Values returns a slice of all values present in the map at the
// moment of the call. It runs in O(n) time.
//
// The result is a point-in-time approximation. Concurrent stores and
// deletes may cause the slice to include values that have since been
// removed, or to omit values that were added during traversal. The
// order of values is undefined, and does not correspond to the order
// returned by Keys.
func (m *SyncMap[K, V]) Values() []V {
	var values []V
	m.syncMap.Range(func(key, value any) bool {
		values = append(values, value.(V))
		return true
	})
	return values
}
