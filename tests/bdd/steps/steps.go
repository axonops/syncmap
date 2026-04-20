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

//go:build bdd

package steps

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/cucumber/godog"

	"github.com/axonops/syncmap"
)

// worldKey is the context key used to store per-scenario World state.
type worldKey struct{}

// World holds per-scenario state. A fresh World is created before each scenario.
type World struct {
	m         *syncmap.SyncMap[string, int]
	prev      int
	found     bool
	ok        bool           // loaded flag from LoadOrStore, or generic bool
	swapped   bool           // CompareAndSwap
	deleted   bool           // CompareAndDelete
	values    []int          // from Values() / captured Range
	keys      []string       // from Keys() / captured Range
	snap      map[string]int // from Map()
	length    int            // from Len()
	rangeHits []string       // ordered sequence of keys visited by Range
}

func newWorld() *World {
	return &World{
		m: &syncmap.SyncMap[string, int]{},
	}
}

func worldFrom(ctx context.Context) *World {
	w, ok := ctx.Value(worldKey{}).(*World)
	if !ok {
		panic("bdd: World missing from context — Before hook did not run")
	}
	return w
}

// Register wires all step definitions into the given scenario context.
func Register(sc *godog.ScenarioContext) {
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		return context.WithValue(ctx, worldKey{}, newWorld()), nil
	})

	// ── Given ──────────────────────────────────────────────────────────────

	sc.Step(`^an empty SyncMap of string to int$`, func(ctx context.Context) error {
		worldFrom(ctx).m = &syncmap.SyncMap[string, int]{}
		return nil
	})

	sc.Step(`^the map contains the following entries$`, func(ctx context.Context, table *godog.Table) error {
		w := worldFrom(ctx)
		for _, row := range table.Rows[1:] { // skip header row
			key := row.Cells[0].Value
			val, err := strconv.Atoi(row.Cells[1].Value)
			if err != nil {
				return fmt.Errorf("invalid value %q in table: %w", row.Cells[1].Value, err)
			}
			w.m.Store(key, val)
		}
		return nil
	})

	sc.Step(`^the key "([^"]*)" has been stored with value (-?\d+)$`, func(ctx context.Context, key string, value int) error {
		worldFrom(ctx).m.Store(key, value)
		return nil
	})

	// ── When ───────────────────────────────────────────────────────────────

	sc.Step(`^I Store key "([^"]*)" with value (-?\d+)$`, func(ctx context.Context, key string, value int) error {
		worldFrom(ctx).m.Store(key, value)
		return nil
	})

	sc.Step(`^I Load key "([^"]*)"$`, func(ctx context.Context, key string) error {
		w := worldFrom(ctx)
		w.prev, w.found = w.m.Load(key)
		return nil
	})

	sc.Step(`^I LoadOrStore key "([^"]*)" with value (-?\d+)$`, func(ctx context.Context, key string, value int) error {
		w := worldFrom(ctx)
		w.prev, w.ok = w.m.LoadOrStore(key, value)
		return nil
	})

	sc.Step(`^I LoadAndDelete key "([^"]*)"$`, func(ctx context.Context, key string) error {
		w := worldFrom(ctx)
		w.prev, w.found = w.m.LoadAndDelete(key)
		return nil
	})

	sc.Step(`^I Delete key "([^"]*)"$`, func(ctx context.Context, key string) error {
		worldFrom(ctx).m.Delete(key)
		return nil
	})

	sc.Step(`^I Swap key "([^"]*)" with value (-?\d+)$`, func(ctx context.Context, key string, value int) error {
		w := worldFrom(ctx)
		w.prev, w.found = w.m.Swap(key, value)
		return nil
	})

	sc.Step(`^I Clear the map$`, func(ctx context.Context) error {
		worldFrom(ctx).m.Clear()
		return nil
	})

	sc.Step(`^I Range all entries$`, func(ctx context.Context) error {
		w := worldFrom(ctx)
		w.rangeHits = nil
		w.m.Range(func(key string, _ int) bool {
			w.rangeHits = append(w.rangeHits, key)
			return true
		})
		return nil
	})

	sc.Step(`^I Range and stop after (\d+) entries$`, func(ctx context.Context, n int) error {
		w := worldFrom(ctx)
		w.rangeHits = nil
		count := 0
		w.m.Range(func(key string, _ int) bool {
			w.rangeHits = append(w.rangeHits, key)
			count++
			return count < n
		})
		return nil
	})

	sc.Step(`^I request Len$`, func(ctx context.Context) error {
		w := worldFrom(ctx)
		w.length = w.m.Len()
		return nil
	})

	sc.Step(`^I request Map$`, func(ctx context.Context) error {
		w := worldFrom(ctx)
		w.snap = w.m.Map()
		return nil
	})

	sc.Step(`^I request Keys$`, func(ctx context.Context) error {
		w := worldFrom(ctx)
		w.keys = w.m.Keys()
		return nil
	})

	sc.Step(`^I request Values$`, func(ctx context.Context) error {
		w := worldFrom(ctx)
		w.values = w.m.Values()
		return nil
	})

	sc.Step(`^I CompareAndSwap key "([^"]*)" from (-?\d+) to (-?\d+)$`, func(ctx context.Context, key string, from, to int) error {
		w := worldFrom(ctx)
		w.swapped = syncmap.CompareAndSwap(w.m, key, from, to)
		return nil
	})

	sc.Step(`^I CompareAndDelete key "([^"]*)" expecting (-?\d+)$`, func(ctx context.Context, key string, old int) error {
		w := worldFrom(ctx)
		w.deleted = syncmap.CompareAndDelete(w.m, key, old)
		return nil
	})

	sc.Step(`^(\d+) goroutines each Store (\d+) keys$`, func(ctx context.Context, goroutines, keysEach int) error {
		w := worldFrom(ctx)
		var wg sync.WaitGroup
		for g := 0; g < goroutines; g++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for k := 0; k < keysEach; k++ {
					key := fmt.Sprintf("g%d-k%d", id, k)
					w.m.Store(key, id*keysEach+k)
				}
			}(g)
		}
		wg.Wait()
		return nil
	})

	// ── Then ───────────────────────────────────────────────────────────────

	sc.Step(`^the returned value is (-?\d+)$`, func(ctx context.Context, expected int) error {
		w := worldFrom(ctx)
		if w.prev != expected {
			return fmt.Errorf("expected returned value %d, got %d", expected, w.prev)
		}
		return nil
	})

	sc.Step(`^the returned value is the zero value$`, func(ctx context.Context) error {
		w := worldFrom(ctx)
		if w.prev != 0 {
			return fmt.Errorf("expected zero value, got %d", w.prev)
		}
		return nil
	})

	sc.Step(`^the found flag is (true|false)$`, func(ctx context.Context, flag string) error {
		w := worldFrom(ctx)
		expected := flag == "true"
		if w.found != expected {
			return fmt.Errorf("expected found=%v, got %v", expected, w.found)
		}
		return nil
	})

	sc.Step(`^the loaded flag is (true|false)$`, func(ctx context.Context, flag string) error {
		w := worldFrom(ctx)
		expected := flag == "true"
		if w.ok != expected {
			return fmt.Errorf("expected loaded=%v, got %v", expected, w.ok)
		}
		return nil
	})

	sc.Step(`^the swapped flag is (true|false)$`, func(ctx context.Context, flag string) error {
		w := worldFrom(ctx)
		expected := flag == "true"
		if w.swapped != expected {
			return fmt.Errorf("expected swapped=%v, got %v", expected, w.swapped)
		}
		return nil
	})

	sc.Step(`^the deleted flag is (true|false)$`, func(ctx context.Context, flag string) error {
		w := worldFrom(ctx)
		expected := flag == "true"
		if w.deleted != expected {
			return fmt.Errorf("expected deleted=%v, got %v", expected, w.deleted)
		}
		return nil
	})

	sc.Step(`^Len returns (\d+)$`, func(ctx context.Context, expected int) error {
		w := worldFrom(ctx)
		if w.length != expected {
			return fmt.Errorf("expected Len %d, got %d", expected, w.length)
		}
		return nil
	})

	sc.Step(`^Len equals (\d+)$`, func(ctx context.Context, expected int) error {
		w := worldFrom(ctx)
		got := w.m.Len()
		if got != expected {
			return fmt.Errorf("expected Len %d, got %d", expected, got)
		}
		return nil
	})

	sc.Step(`^the captured values contain (-?\d+)$`, func(ctx context.Context, expected int) error {
		w := worldFrom(ctx)
		for _, v := range w.values {
			if v == expected {
				return nil
			}
		}
		return fmt.Errorf("captured values %v do not contain %d", w.values, expected)
	})

	sc.Step(`^the captured values length equals (\d+)$`, func(ctx context.Context, expected int) error {
		w := worldFrom(ctx)
		got := len(w.values)
		if got != expected {
			return fmt.Errorf("expected captured values length %d, got %d", expected, got)
		}
		return nil
	})

	sc.Step(`^the captured keys contain "([^"]*)"$`, func(ctx context.Context, expected string) error {
		w := worldFrom(ctx)
		for _, k := range w.keys {
			if k == expected {
				return nil
			}
		}
		return fmt.Errorf("captured keys %v do not contain %q", w.keys, expected)
	})

	sc.Step(`^the captured keys length equals (\d+)$`, func(ctx context.Context, expected int) error {
		w := worldFrom(ctx)
		got := len(w.keys)
		if got != expected {
			return fmt.Errorf("expected captured keys length %d, got %d", expected, got)
		}
		return nil
	})

	sc.Step(`^the snapshot length equals (\d+)$`, func(ctx context.Context, expected int) error {
		w := worldFrom(ctx)
		got := len(w.snap)
		if got != expected {
			return fmt.Errorf("expected snapshot length %d, got %d", expected, got)
		}
		return nil
	})

	sc.Step(`^the snapshot contains key "([^"]*)" with value (-?\d+)$`, func(ctx context.Context, key string, expected int) error {
		w := worldFrom(ctx)
		got, ok := w.snap[key]
		if !ok {
			return fmt.Errorf("snapshot does not contain key %q", key)
		}
		if got != expected {
			return fmt.Errorf("snapshot[%q] = %d, want %d", key, got, expected)
		}
		return nil
	})

	sc.Step(`^Range visited exactly (\d+) entries$`, func(ctx context.Context, expected int) error {
		w := worldFrom(ctx)
		got := len(w.rangeHits)
		if got != expected {
			return fmt.Errorf("expected Range to visit %d entries, visited %d", expected, got)
		}
		return nil
	})

	sc.Step(`^the map does not contain key "([^"]*)"$`, func(ctx context.Context, key string) error {
		w := worldFrom(ctx)
		_, ok := w.m.Load(key)
		if ok {
			return fmt.Errorf("map unexpectedly contains key %q", key)
		}
		return nil
	})

	sc.Step(`^the map contains key "([^"]*)" with value (-?\d+)$`, func(ctx context.Context, key string, expected int) error {
		w := worldFrom(ctx)
		got, ok := w.m.Load(key)
		if !ok {
			return fmt.Errorf("map does not contain key %q", key)
		}
		if got != expected {
			return fmt.Errorf("map[%q] = %d, want %d", key, got, expected)
		}
		return nil
	})

	sc.Step(`^no panic occurs$`, func(_ context.Context) error {
		// No-op: Go tests panic automatically on a real panic.
		// This step documents the contract that no panic should occur.
		return nil
	})
}
