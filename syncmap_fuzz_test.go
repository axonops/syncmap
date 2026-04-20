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
	"strings"
	"sync"
	"testing"

	"github.com/axonops/syncmap"
)

// FuzzLoadStore verifies that a value stored under a key is always retrievable
// with the correct value and found==true. The fuzz engine varies both the key
// and the value.
func FuzzLoadStore(f *testing.F) {
	// Seed corpus
	f.Add("", 0)
	f.Add("k", 1)
	f.Add("\x00key", -1)
	f.Add("ü", 2147483647)
	f.Add("a very long key "+strings.Repeat("x", 128), -42)

	f.Fuzz(func(t *testing.T, k string, v int) {
		var m syncmap.SyncMap[string, int]
		m.Store(k, v)
		got, ok := m.Load(k)
		if !ok {
			t.Fatalf("Load(%q): expected found=true after Store, got false", k)
		}
		if got != v {
			t.Errorf("Load(%q): expected %d, got %d", k, v, got)
		}
	})
}

// FuzzConcurrent exercises concurrent Load, Store, Delete, and LoadOrStore
// operations driven by arbitrary byte sequences. It must not panic and must
// be clean under -race.
//
// No ordering assertions are made; the test fails only on panic or a data race
// detected by the race detector.
func FuzzConcurrent(f *testing.F) {
	// Seed corpus
	f.Add([]byte(""))
	f.Add([]byte("\x00"))
	f.Add([]byte("\x01\x02\x03\x04"))
	f.Add([]byte("\xff\xfe\xfd\x00\x01\x02\x03"))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}

		var m syncmap.SyncMap[string, int]

		// Distribute the byte slice across 4 goroutines. Each goroutine processes
		// its own quarter of the data so the workload genuinely exercises
		// concurrent access without ordering assumptions.
		const numGoroutines = 4
		chunkSize := (len(data) + numGoroutines - 1) / numGoroutines

		var wg sync.WaitGroup
		for g := 0; g < numGoroutines; g++ {
			start := g * chunkSize
			if start >= len(data) {
				break
			}
			end := start + chunkSize
			if end > len(data) {
				end = len(data)
			}
			chunk := data[start:end]

			wg.Add(1)
			go func(chunk []byte) {
				defer wg.Done()
				for _, b := range chunk {
					op := b % 4
					key := strconv.Itoa(int(b) % 8)
					value := int(b)
					switch op {
					case 0:
						m.Load(key)
					case 1:
						m.Store(key, value)
					case 2:
						m.Delete(key)
					case 3:
						m.LoadOrStore(key, value)
					}
				}
			}(chunk)
		}

		wg.Wait()
	})
}
