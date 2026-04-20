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
// landing in issue #14 (CompareAndSwap / CompareAndDelete). A full
// benchmark set covering every public method, plus a raw-sync.Map
// overhead comparison and a committed bench.txt baseline, is owned by
// issue #15.

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
