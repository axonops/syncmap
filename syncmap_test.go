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
	"sort"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/axonops/syncmap"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestLoad(t *testing.T) {
	t.Parallel()

	t.Run("present", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, int]{}
		m.Store("k", 42)
		v, ok := m.Load("k")
		assert.True(t, ok)
		assert.Equal(t, 42, v)
	})

	t.Run("missing", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, int]{}
		v, ok := m.Load("absent")
		assert.False(t, ok)
		assert.Zero(t, v)
	})

	t.Run("zero_value_stored", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, int]{}
		m.Store("z", 0)
		v, ok := m.Load("z")
		assert.True(t, ok)
		assert.Equal(t, 0, v)
	})

	t.Run("empty_key", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, string]{}
		m.Store("", "empty")
		v, ok := m.Load("")
		assert.True(t, ok)
		assert.Equal(t, "empty", v)
	})

	t.Run("zero_map", func(t *testing.T) {
		t.Parallel()
		var m syncmap.SyncMap[string, int]
		v, ok := m.Load("anything")
		assert.False(t, ok)
		assert.Zero(t, v)
	})
}

func TestStore(t *testing.T) {
	t.Parallel()

	t.Run("new_key", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, int]{}
		m.Store("a", 1)
		v, ok := m.Load("a")
		assert.True(t, ok)
		assert.Equal(t, 1, v)
	})

	t.Run("overwrite", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, int]{}
		m.Store("a", 1)
		m.Store("a", 2)
		v, ok := m.Load("a")
		assert.True(t, ok)
		assert.Equal(t, 2, v)
	})

	t.Run("empty_key", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, string]{}
		m.Store("", "val")
		v, ok := m.Load("")
		assert.True(t, ok)
		assert.Equal(t, "val", v)
	})

	t.Run("pointer_value", func(t *testing.T) {
		t.Parallel()
		type payload struct{ n int }
		m := &syncmap.SyncMap[string, *payload]{}
		p := &payload{n: 7}
		m.Store("ptr", p)
		got, ok := m.Load("ptr")
		assert.True(t, ok)
		assert.Same(t, p, got)
	})
}

func TestLoadOrStore(t *testing.T) {
	t.Parallel()

	t.Run("absent_stores", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, int]{}
		actual, loaded := m.LoadOrStore("k", 10)
		assert.False(t, loaded)
		assert.Equal(t, 10, actual)
		v, ok := m.Load("k")
		require.True(t, ok)
		assert.Equal(t, 10, v)
	})

	t.Run("present_loads", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, int]{}
		m.Store("k", 99)
		actual, loaded := m.LoadOrStore("k", 0)
		assert.True(t, loaded)
		assert.Equal(t, 99, actual)
	})

	t.Run("zero_value_loadable", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, int]{}
		m.Store("k", 0)
		actual, loaded := m.LoadOrStore("k", 1)
		assert.True(t, loaded)
		assert.Equal(t, 0, actual)
	})
}

func TestLoadAndDelete(t *testing.T) {
	t.Parallel()

	t.Run("present", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, int]{}
		m.Store("k", 55)
		v, loaded := m.LoadAndDelete("k")
		assert.True(t, loaded)
		assert.Equal(t, 55, v)
		_, ok := m.Load("k")
		assert.False(t, ok)
	})

	t.Run("missing", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, int]{}
		v, loaded := m.LoadAndDelete("absent")
		assert.False(t, loaded)
		assert.Zero(t, v)
	})

	t.Run("zero_value_distinguished_from_missing", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, int]{}
		m.Store("k", 0)
		v, loaded := m.LoadAndDelete("k")
		assert.True(t, loaded)
		assert.Equal(t, 0, v)
		_, stillPresent := m.Load("k")
		assert.False(t, stillPresent)
	})
}

func TestDelete(t *testing.T) {
	t.Parallel()

	t.Run("existing", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, int]{}
		m.Store("k", 1)
		m.Delete("k")
		_, ok := m.Load("k")
		assert.False(t, ok)
	})

	t.Run("missing_noop", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, int]{}
		assert.NotPanics(t, func() { m.Delete("never-existed") })
	})

	t.Run("double_delete", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, int]{}
		m.Store("k", 1)
		m.Delete("k")
		assert.NotPanics(t, func() { m.Delete("k") })
		_, ok := m.Load("k")
		assert.False(t, ok)
	})
}

func TestRange(t *testing.T) {
	t.Parallel()

	t.Run("all_visited", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[int, int]{}
		for i := 0; i < 5; i++ {
			m.Store(i, i*10)
		}
		seen := make(map[int]int)
		m.Range(func(k, v int) bool {
			seen[k] = v
			return true
		})
		assert.Equal(t, 5, len(seen))
		for i := 0; i < 5; i++ {
			assert.Equal(t, i*10, seen[i])
		}
	})

	t.Run("early_return_stops", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[int, int]{}
		for i := 0; i < 10; i++ {
			m.Store(i, i)
		}
		count := 0
		m.Range(func(k, v int) bool {
			count++
			return count < 3
		})
		assert.Equal(t, 3, count)
	})

	t.Run("empty_map_no_calls", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, string]{}
		calls := 0
		m.Range(func(k, v string) bool {
			calls++
			return true
		})
		assert.Equal(t, 0, calls)
	})
}

func TestLen(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, int]{}
		assert.Equal(t, 0, m.Len())
	})

	t.Run("after_store", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, int]{}
		m.Store("a", 1)
		m.Store("b", 2)
		m.Store("c", 3)
		assert.Equal(t, 3, m.Len())
	})

	t.Run("after_delete", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, int]{}
		m.Store("a", 1)
		m.Store("b", 2)
		m.Delete("a")
		assert.Equal(t, 1, m.Len())
	})
}

func TestMap(t *testing.T) {
	t.Parallel()

	t.Run("snapshot_independence", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, int]{}
		m.Store("x", 1)
		snap := m.Map()
		require.NotNil(t, snap)
		snap["x"] = 999
		snap["y"] = 42
		v, ok := m.Load("x")
		assert.True(t, ok)
		assert.Equal(t, 1, v)
		_, ok = m.Load("y")
		assert.False(t, ok)
	})

	t.Run("empty_returns_non_nil_empty", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, int]{}
		snap := m.Map()
		assert.NotNil(t, snap)
		assert.Empty(t, snap)
	})

	t.Run("matches_contents", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, int]{}
		m.Store("a", 1)
		m.Store("b", 2)
		snap := m.Map()
		assert.Equal(t, map[string]int{"a": 1, "b": 2}, snap)
	})
}

func TestKeys(t *testing.T) {
	t.Parallel()

	t.Run("empty_is_empty", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, int]{}
		keys := m.Keys()
		assert.Empty(t, keys)
	})

	t.Run("matches", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, int]{}
		m.Store("a", 1)
		m.Store("b", 2)
		m.Store("c", 3)
		keys := m.Keys()
		sort.Strings(keys)
		assert.Equal(t, []string{"a", "b", "c"}, keys)
	})

	t.Run("length_equals_len", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, int]{}
		m.Store("x", 10)
		m.Store("y", 20)
		assert.Equal(t, m.Len(), len(m.Keys()))
	})
}

func TestItems(t *testing.T) {
	t.Parallel()

	t.Run("empty_is_empty", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, int]{}
		items := m.Items()
		assert.Empty(t, items)
	})

	t.Run("matches", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, int]{}
		m.Store("a", 1)
		m.Store("b", 2)
		m.Store("c", 3)
		items := m.Items()
		sort.Ints(items)
		assert.Equal(t, []int{1, 2, 3}, items)
	})

	t.Run("length_equals_len", func(t *testing.T) {
		t.Parallel()
		m := &syncmap.SyncMap[string, int]{}
		m.Store("x", 10)
		m.Store("y", 20)
		assert.Equal(t, m.Len(), len(m.Items()))
	})
}

func TestZeroValueUsable(t *testing.T) {
	t.Parallel()
	var sm syncmap.SyncMap[string, int]
	sm.Store("k", 77)
	v, ok := sm.Load("k")
	assert.True(t, ok)
	assert.Equal(t, 77, v)
}

func TestConcurrentWritersReaders(t *testing.T) {
	t.Parallel()

	const writers = 16
	const readers = 8
	const opsEach = 1000

	m := &syncmap.SyncMap[int, int]{}

	// Track which keys each writer committed so we can verify the final state.
	// Each writer i owns keys in the range [i*opsEach, (i+1)*opsEach).
	var wg sync.WaitGroup

	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			base := id * opsEach
			for j := 0; j < opsEach; j++ {
				m.Store(base+j, id)
			}
		}(w)
	}

	for r := 0; r < readers; r++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsEach; j++ {
				key := (id*opsEach + j) % (writers * opsEach)
				if v, ok := m.Load(key); ok {
					assert.GreaterOrEqual(t, v, 0)
					assert.Less(t, v, writers)
				}
			}
		}(r)
	}

	wg.Wait()

	assert.Equal(t, writers*opsEach, m.Len())
}

func TestLoadOrStoreContention(t *testing.T) {
	t.Parallel()

	const goroutines = 100

	m := &syncmap.SyncMap[string, int]{}
	var stored atomic.Int32
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, loaded := m.LoadOrStore("contended", 1)
			if !loaded {
				stored.Add(1)
			}
		}()
	}

	wg.Wait()

	assert.Equal(t, int32(1), stored.Load(), "exactly one goroutine should win the store")
}

func TestRangeDuringWrites(t *testing.T) {
	t.Parallel()

	const keyCount = 100
	const writerCount = 8

	m := &syncmap.SyncMap[int, int]{}
	for i := 0; i < keyCount; i++ {
		m.Store(i, i)
	}

	var wg sync.WaitGroup

	// Writers continuously overwrite keys across the whole key space while
	// Range runs. Each writer cycles through all keys so the concurrent
	// workload actually contends with the ranger.
	var stop atomic.Int32
	for w := 0; w < writerCount; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for stop.Load() == 0 {
				for j := 0; j < keyCount; j++ {
					m.Store(j, id)
				}
			}
		}(w)
	}

	// A single Range pass; per sync.Map contract it may observe any subset
	// of keys that were present at or after the Range started.
	seen := make(map[int]struct{})
	m.Range(func(k, v int) bool {
		seen[k] = struct{}{}
		return true
	})

	stop.Store(1)
	wg.Wait()

	// All observed keys must be within the valid key space [0, keyCount).
	for k := range seen {
		assert.GreaterOrEqual(t, k, 0)
		assert.Less(t, k, keyCount)
	}
}

func TestDeleteDuringRange(t *testing.T) {
	t.Parallel()

	const keyCount = 50

	m := &syncmap.SyncMap[int, int]{}
	for i := 0; i < keyCount; i++ {
		m.Store(i, i)
	}

	var wg sync.WaitGroup

	// Deleter removes all keys concurrently with the Range below.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < keyCount; i++ {
			m.Delete(i)
		}
	}()

	// Range must not panic regardless of concurrent deletes. Per sync.Map
	// docs, a key deleted during Range may appear zero or one times — both
	// outcomes are correct; we assert only on absence of panic and value
	// validity.
	assert.NotPanics(t, func() {
		m.Range(func(k, v int) bool {
			// Value must be in the valid range if observed.
			assert.GreaterOrEqual(t, k, 0)
			assert.Less(t, k, keyCount)
			assert.Equal(t, k, v)
			return true
		})
	})

	wg.Wait()
}
