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
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/axonops/syncmap"
)

// Benchmark suite for syncmap.
//
// Scope: every public method plus overhead pairs comparing the generic
// wrapper against raw sync.Map. The committed bench.txt baseline is the
// artefact this file produces; benchstat-regression-guard in CI compares
// a fresh run against that baseline on every PR.
//
// Regenerate the baseline with `make bench > bench.txt` (strip the
// trailing `PASS` / `ok` lines and ANSI escapes before committing) and
// land the update in the same PR as any performance-affecting change.

func BenchmarkCompareAndSwap(b *testing.B) {
	b.ReportAllocs()
	var m syncmap.SyncMap[string, int]
	m.Store("k", 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Alternate old→new so every call performs a real swap.
		syncmap.CompareAndSwap(&m, "k", i, i+1)
	}
}

func BenchmarkCompareAndSwapMismatch(b *testing.B) {
	b.ReportAllocs()
	var m syncmap.SyncMap[string, int]
	m.Store("k", 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Old never matches — exercises the fast-reject path.
		syncmap.CompareAndSwap(&m, "k", -1, i)
	}
}

func BenchmarkCompareAndDelete(b *testing.B) {
	b.ReportAllocs()
	var m syncmap.SyncMap[string, int]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Store("k", 0)
		syncmap.CompareAndDelete(&m, "k", 0)
	}
}

func BenchmarkCompareAndSwapParallel(b *testing.B) {
	b.ReportAllocs()
	var m syncmap.SyncMap[string, int]
	m.Store("k", 0)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// Most attempts will fail (only one goroutine's old
			// matches at any moment), which is the realistic
			// contention pattern.
			syncmap.CompareAndSwap(&m, "k", i, i+1)
			i++
		}
	})
}

func BenchmarkSwap(b *testing.B) {
	b.ReportAllocs()
	var m syncmap.SyncMap[string, int]
	m.Store("k", 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Swap("k", i)
	}
}

func BenchmarkSwapAbsent(b *testing.B) {
	b.ReportAllocs()
	// int keys avoid the string-hash cost so the measured overhead
	// is dominated by the !loaded guard and sync.Map's fresh-entry path.
	var m syncmap.SyncMap[int, int]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Fresh key each iteration — exercises the !loaded guard.
		m.Swap(i, i)
	}
}

func BenchmarkSwapParallel(b *testing.B) {
	b.ReportAllocs()
	var m syncmap.SyncMap[string, int]
	m.Store("k", 0)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Swap("k", i)
			i++
		}
	})
}

func BenchmarkClear(b *testing.B) {
	b.ReportAllocs()
	var m syncmap.SyncMap[string, int]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Cost includes one Store per iteration — Clear on an empty map
		// is meaningless, so this measures the combined cost. Isolate
		// via pprof if the Clear fraction needs to be teased apart.
		m.Store("k", 0)
		m.Clear()
	}
}

func BenchmarkClearParallel(b *testing.B) {
	b.ReportAllocs()
	// Each goroutine owns its own map so Clear can race with Stores
	// without invalidating per-iteration semantics. Concurrent Clear
	// on a shared map is a legitimate pattern but the result is less
	// informative (you can't reason about what any iteration "did").
	b.RunParallel(func(pb *testing.PB) {
		var m syncmap.SyncMap[string, int]
		for pb.Next() {
			m.Store("k", 0)
			m.Clear()
		}
	})
}

// -----------------------------------------------------------------------------
// Per-method benchmarks — Load, Store, LoadOrStore, LoadAndDelete, Delete,
// Range, Len, Map, Keys, Values.
// -----------------------------------------------------------------------------

func BenchmarkLoad(b *testing.B) {
	b.ReportAllocs()
	var m syncmap.SyncMap[string, int]
	m.Store("k", 42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.Load("k")
	}
}

func BenchmarkLoadMiss(b *testing.B) {
	b.ReportAllocs()
	var m syncmap.SyncMap[string, int]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.Load("absent")
	}
}

func BenchmarkStore(b *testing.B) {
	b.ReportAllocs()
	var m syncmap.SyncMap[string, int]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Store("k", i)
	}
}

func BenchmarkLoadOrStoreLoaded(b *testing.B) {
	b.ReportAllocs()
	var m syncmap.SyncMap[string, int]
	m.Store("k", 42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.LoadOrStore("k", 0)
	}
}

func BenchmarkLoadOrStoreStored(b *testing.B) {
	b.ReportAllocs()
	var m syncmap.SyncMap[int, int]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.LoadOrStore(i, i)
	}
}

func BenchmarkLoadAndDelete(b *testing.B) {
	b.ReportAllocs()
	var m syncmap.SyncMap[int, int]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Store(i, i)
		_, _ = m.LoadAndDelete(i)
	}
}

func BenchmarkDelete(b *testing.B) {
	b.ReportAllocs()
	var m syncmap.SyncMap[int, int]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Store(i, i)
		m.Delete(i)
	}
}

// Size-parameterised benchmarks for O(n) helpers.

const benchMapSize = 1000

func seedMap(n int) *syncmap.SyncMap[string, int] {
	m := &syncmap.SyncMap[string, int]{}
	for i := 0; i < n; i++ {
		m.Store(strconv.Itoa(i), i)
	}
	return m
}

func BenchmarkRange(b *testing.B) {
	b.ReportAllocs()
	m := seedMap(benchMapSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Range(func(k string, v int) bool { return true })
	}
}

func BenchmarkLen(b *testing.B) {
	b.ReportAllocs()
	m := seedMap(benchMapSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Len()
	}
}

func BenchmarkMap(b *testing.B) {
	b.ReportAllocs()
	m := seedMap(benchMapSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Map()
	}
}

func BenchmarkKeys(b *testing.B) {
	b.ReportAllocs()
	m := seedMap(benchMapSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Keys()
	}
}

func BenchmarkValues(b *testing.B) {
	b.ReportAllocs()
	m := seedMap(benchMapSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Values()
	}
}

// -----------------------------------------------------------------------------
// Concurrent access pattern — realistic 90% read / 10% write mix.
// -----------------------------------------------------------------------------

func BenchmarkConcurrentReadWrite(b *testing.B) {
	b.ReportAllocs()
	m := seedMap(benchMapSize)

	// Pre-compute the key pool so the timed loop doesn't allocate on
	// strconv.Itoa every iteration — otherwise the allocs/op signal is
	// dominated by the benchmark harness rather than the map.
	keys := make([]string, benchMapSize)
	for i := range keys {
		keys[i] = strconv.Itoa(i)
	}
	var counter atomic.Int64

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := counter.Add(1)
			key := keys[int(n%benchMapSize)]
			if n%10 == 0 {
				m.Store(key, int(n))
			} else {
				_, _ = m.Load(key)
			}
		}
	})
}

// -----------------------------------------------------------------------------
// Overhead pairs vs raw sync.Map — measures the wrapper cost beyond stdlib.
// Both sides perform the same operations with the same workload. For the
// Delete and LoadAndDelete pairs, each iteration includes a Store so the
// method under test has something to operate on; the pair compares the
// generic wrapper's Store+Delete cost against the raw sync.Map's
// Store+Delete cost — any delta is wrapper overhead, not the absolute
// cost of the named operation.
// -----------------------------------------------------------------------------

func BenchmarkOverhead_LoadSyncMap(b *testing.B) {
	b.ReportAllocs()
	var m syncmap.SyncMap[string, int]
	m.Store("k", 42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.Load("k")
	}
}

func BenchmarkOverhead_LoadRawSyncMap(b *testing.B) {
	b.ReportAllocs()
	var m sync.Map
	m.Store("k", 42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if v, ok := m.Load("k"); ok {
			_ = v.(int)
		}
	}
}

func BenchmarkOverhead_StoreSyncMap(b *testing.B) {
	b.ReportAllocs()
	var m syncmap.SyncMap[string, int]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Store("k", i)
	}
}

func BenchmarkOverhead_StoreRawSyncMap(b *testing.B) {
	b.ReportAllocs()
	var m sync.Map
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Store("k", i)
	}
}

func BenchmarkOverhead_LoadOrStoreSyncMap(b *testing.B) {
	b.ReportAllocs()
	var m syncmap.SyncMap[int, int]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.LoadOrStore(i, i)
	}
}

func BenchmarkOverhead_LoadOrStoreRawSyncMap(b *testing.B) {
	b.ReportAllocs()
	var m sync.Map
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if v, _ := m.LoadOrStore(i, i); v != nil {
			_ = v.(int)
		}
	}
}

func BenchmarkOverhead_DeleteSyncMap(b *testing.B) {
	b.ReportAllocs()
	var m syncmap.SyncMap[int, int]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Store(i, i)
		m.Delete(i)
	}
}

func BenchmarkOverhead_DeleteRawSyncMap(b *testing.B) {
	b.ReportAllocs()
	var m sync.Map
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Store(i, i)
		m.Delete(i)
	}
}

func BenchmarkOverhead_LoadAndDeleteSyncMap(b *testing.B) {
	b.ReportAllocs()
	var m syncmap.SyncMap[int, int]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Store(i, i)
		_, _ = m.LoadAndDelete(i)
	}
}

func BenchmarkOverhead_LoadAndDeleteRawSyncMap(b *testing.B) {
	b.ReportAllocs()
	var m sync.Map
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Store(i, i)
		if v, loaded := m.LoadAndDelete(i); loaded {
			_ = v.(int)
		}
	}
}
