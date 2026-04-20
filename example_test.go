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

package syncmap_test

import (
	"fmt"
	"sort"

	"github.com/axonops/syncmap"
)

// Range, Keys, Values, and Map make no order guarantee. Every Example
// that iterates over the map sorts its output before printing so the
// // Output: blocks are deterministic under `go test`.

func ExampleSyncMap() {
	var m syncmap.SyncMap[string, int]

	m.Store("hits", 1)
	m.Store("misses", 0)

	v, ok := m.Load("hits")
	fmt.Println(v, ok)
	// Output: 1 true
}

func ExampleSyncMap_Load() {
	var m syncmap.SyncMap[string, int]
	m.Store("answer", 42)

	v, ok := m.Load("answer")
	fmt.Println(v, ok)

	v, ok = m.Load("missing")
	fmt.Println(v, ok)
	// Output:
	// 42 true
	// 0 false
}

func ExampleSyncMap_Store() {
	var m syncmap.SyncMap[string, string]
	m.Store("env", "prod")
	m.Store("env", "staging") // overwrites

	v, _ := m.Load("env")
	fmt.Println(v)
	// Output: staging
}

func ExampleSyncMap_LoadOrStore() {
	var m syncmap.SyncMap[string, int]

	v1, loaded1 := m.LoadOrStore("k", 1)
	v2, loaded2 := m.LoadOrStore("k", 2)

	fmt.Println(v1, loaded1)
	fmt.Println(v2, loaded2)
	// Output:
	// 1 false
	// 1 true
}

func ExampleSyncMap_LoadAndDelete() {
	var m syncmap.SyncMap[string, int]
	m.Store("k", 7)

	v, loaded := m.LoadAndDelete("k")
	fmt.Println(v, loaded)

	v, loaded = m.LoadAndDelete("k")
	fmt.Println(v, loaded)
	// Output:
	// 7 true
	// 0 false
}

func ExampleSyncMap_Delete() {
	var m syncmap.SyncMap[string, int]
	m.Store("k", 1)
	m.Delete("k")

	_, ok := m.Load("k")
	fmt.Println(ok)
	// Output: false
}

func ExampleSyncMap_Swap() {
	var m syncmap.SyncMap[string, int]

	previous, loaded := m.Swap("k", 1)
	fmt.Println(previous, loaded)

	previous, loaded = m.Swap("k", 2)
	fmt.Println(previous, loaded)
	// Output:
	// 0 false
	// 1 true
}

func ExampleSyncMap_Clear() {
	var m syncmap.SyncMap[string, int]
	m.Store("a", 1)
	m.Store("b", 2)

	m.Clear()
	fmt.Println(m.Len())
	// Output: 0
}

func ExampleSyncMap_Range() {
	var m syncmap.SyncMap[string, int]
	m.Store("a", 1)
	m.Store("b", 2)
	m.Store("c", 3)

	var keys []string
	m.Range(func(k string, v int) bool {
		keys = append(keys, fmt.Sprintf("%s=%d", k, v))
		return true
	})
	sort.Strings(keys)
	for _, entry := range keys {
		fmt.Println(entry)
	}
	// Output:
	// a=1
	// b=2
	// c=3
}

func ExampleSyncMap_Len() {
	var m syncmap.SyncMap[string, int]
	m.Store("a", 1)
	m.Store("b", 2)
	fmt.Println(m.Len())
	// Output: 2
}

func ExampleSyncMap_Map() {
	var m syncmap.SyncMap[string, int]
	m.Store("a", 1)
	m.Store("b", 2)

	snap := m.Map()
	keys := make([]string, 0, len(snap))
	for k := range snap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("%s=%d\n", k, snap[k])
	}
	// Output:
	// a=1
	// b=2
}

func ExampleSyncMap_Keys() {
	var m syncmap.SyncMap[string, int]
	m.Store("a", 1)
	m.Store("b", 2)

	keys := m.Keys()
	sort.Strings(keys)
	fmt.Println(keys)
	// Output: [a b]
}

func ExampleSyncMap_Values() {
	var m syncmap.SyncMap[string, int]
	m.Store("a", 1)
	m.Store("b", 2)

	values := m.Values()
	sort.Ints(values)
	fmt.Println(values)
	// Output: [1 2]
}

func ExampleCompareAndSwap() {
	var m syncmap.SyncMap[string, int]
	m.Store("k", 1)

	swapped := syncmap.CompareAndSwap(&m, "k", 1, 2)
	fmt.Println(swapped)

	swapped = syncmap.CompareAndSwap(&m, "k", 1, 3)
	fmt.Println(swapped)
	// Output:
	// true
	// false
}

func ExampleCompareAndDelete() {
	var m syncmap.SyncMap[string, int]
	m.Store("k", 1)

	deleted := syncmap.CompareAndDelete(&m, "k", 2)
	fmt.Println(deleted)

	deleted = syncmap.CompareAndDelete(&m, "k", 1)
	fmt.Println(deleted)
	// Output:
	// false
	// true
}
