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
	"testing"

	"github.com/axonops/syncmap"
)

// This file seeds the benchmark suite with coverage for the functions
// landing in issues #13 (Swap, Clear) and #14 (CompareAndSwap,
// CompareAndDelete). A full benchmark set covering every public
// method, plus a raw-sync.Map overhead comparison and a committed
// bench.txt baseline, is owned by issue #15.

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
