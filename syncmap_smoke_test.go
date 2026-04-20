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

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

// Some basic smoke tests, not very comprehensive

func TestSyncMapGeneric(t *testing.T) {
	sm := &SyncMap[string, string]{}
	var wg sync.WaitGroup
	// Run 11 goroutines in parallel each adding 10 items
	for i := 0; i <= 100; i += 10 {
		wg.Add(1)
		go func(offset int) {
			for v := offset; v < offset+10; v++ {
				sm.Store(fmt.Sprintf("key %d", v), fmt.Sprintf("value %d", v))
			}
			wg.Done()
		}(i)
	}
	wg.Wait()

	// Display and verify the contents of the map
	seenNums := make(map[int]struct{})
	realLen := 0
	sm.Range(func(k string, v string) bool {
		fmt.Printf("%s : %s\n", k, v)
		n, _ := strconv.Atoi(k[4:])
		if n < 0 || n >= 110 {
			t.Errorf("Unexpected number in map: %d", n)
		}
		seenNums[n] = struct{}{}
		realLen++
		return true
	})

	expectedLen := 110
	smLen := sm.Len()
	if smLen != expectedLen {
		t.Errorf("Len() does not match expected length. Expected=%d, Actual=%d", expectedLen, smLen)
	}
	if realLen != expectedLen {
		t.Errorf("Counted length does not match expected length. Expected=%d, Actual=%d", expectedLen, realLen)
	}
	for i := 0; i < 110; i++ {
		_, ok := seenNums[i]
		if !ok {
			t.Errorf("Number %d not found in the map", i)
		}
	}

	fmt.Printf("Total values: %d\n", sm.Len())
}

// Concurrent read and write test

func TestSyncMapReadWrite(t *testing.T) {
	sm := &SyncMap[string, string]{}
	var wg sync.WaitGroup
	// Writers
	for i := 0; i <= 100; i += 10 {
		wg.Add(1)
		go func(offset int) {
			defer wg.Done()
			for k := 0; k < 10; k++ {
				v := k + offset
				ks := fmt.Sprintf("key %d", k)
				vs := fmt.Sprintf("value %d", v)
				sm.Store(ks, vs)
				//fmt.Printf("writer %d set %s = %s\n", offset, ks, vs)
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	// Readers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for n := 0; n < 10; n++ {
				//l := sm.Len()
				k := fmt.Sprintf("key %d", rand.Intn(10))
				v, ok := sm.Load(k)
				if ok {
					fmt.Printf("Reader %d: %s = %s\n", id, k, v)
				} else {
					fmt.Printf("Reader %d: %s not set\n", id, k)
				}
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	fmt.Println()
	sm.Range(func(k string, v string) bool {
		fmt.Printf("%s : %s\n", k, v)
		return true
	})
	fmt.Printf("Total values: %d\n", sm.Len())
}
