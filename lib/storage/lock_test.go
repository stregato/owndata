package storage

import (
	"sync"
	"testing"
	"time"

	"github.com/stregato/mio/lib/core"
)

func TestLock(t *testing.T) {
	store := NewTestStore("local")
	dir := "testDir"
	lockType := "lockTest"

	release, err := Lock(store, dir, lockType, 0) // Extend timeout as needed.
	core.TestErr(t, err, "Failed to acquire lock: %v")
	core.Assert(t, release != nil, "Lock not acquired")

	close(release)
}

func TestLockHighConcurrency(t *testing.T) {
	const concurrentGoroutines = 50
	const workDuration = 10 * time.Millisecond

	// A shared resource to demonstrate the lock's effectiveness.
	var counter int
	store := NewTestStore("local")
	dir := "testDir"
	lockType := "concurrencyTest"

	var wg sync.WaitGroup
	wg.Add(concurrentGoroutines)

	for i := 0; i < concurrentGoroutines; i++ {
		go func(i int) {
			defer wg.Done()

			release, err := Lock(store, dir, lockType, 20*time.Second) // Extend timeout as needed.
			core.TestErr(t, err, "Failed to acquire lock: %v")
			core.Assert(t, release == nil, "Did not acquire lock for goroutine %d")

			// Simulate work by modifying the shared resource under the lock.
			counter++
			time.Sleep(workDuration) // Simulate work duration.

			// Release the lock.
			close(release)
		}(i)
	}

	wg.Wait()

	// The counter should be equal to the number of goroutines that successfully acquired the lock.
	// Note: This assertion assumes all goroutines eventually acquire the lock before the timeout.
	if counter != concurrentGoroutines {
		t.Errorf("Expected counter to be %d, got %d", concurrentGoroutines, counter)
	}
}
