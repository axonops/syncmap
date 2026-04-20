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

// Package syncmap provides a type-safe, generic wrapper around
// [sync.Map].
//
// The standard [sync.Map] stores keys and values as any, which means
// every load and store requires a type assertion at the call site.
// SyncMap[K, V] moves those assertions inside the wrapper, giving
// callers compile-time type safety with no additional allocations and
// no runtime dependencies beyond the standard library.
//
// # Relationship to sync.Map
//
// SyncMap is a thin layer over sync.Map. It exposes the same set of
// operations — Load, Store, LoadOrStore, LoadAndDelete, Delete, and
// Range — with identical semantics and the same concurrency
// guarantees. Four convenience methods are added on top: Len, Map,
// Keys, and Items. The underlying sync.Map is not exported; use the
// typed methods exclusively.
//
// # When to use SyncMap
//
// sync.Map is optimised for two access patterns: (1) entries are
// written once and read many times, or (2) multiple goroutines each
// operate on disjoint sets of keys. For workloads that do not fit
// either pattern — for example, a cache that is frequently written by
// a single goroutine — a plain map protected by a sync.RWMutex will
// usually perform better.
//
// Use SyncMap (and sync.Map) when:
//   - Many goroutines read the same keys concurrently.
//   - The set of active keys is stable; writes are infrequent.
//   - You want a lock-free path for the common read case.
//
// Use map + sync.RWMutex when:
//   - The write rate is high or unpredictable.
//   - You need snapshot-consistent reads of multiple keys at once.
//   - The map is owned by a single goroutine.
//
// # Thread safety
//
// All methods on SyncMap are safe for concurrent use by multiple
// goroutines without additional locking. This guarantee is inherited
// directly from sync.Map.
//
// # Zero value
//
// The zero value of SyncMap is an empty map ready for use. It must
// not be copied after first use; the same restriction applies as for
// sync.Map and sync.Mutex.
//
// # Quick start
//
//	var m syncmap.SyncMap[string, int]
//
//	m.Store("hits", 1)
//
//	if v, ok := m.Load("hits"); ok {
//		fmt.Println(v) // 1
//	}
//
//	m.Range(func(k string, v int) bool {
//		fmt.Printf("%s=%d\n", k, v)
//		return true
//	})
package syncmap
