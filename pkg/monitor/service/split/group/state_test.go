package group

import (
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Test for deadlock prevention in the Merge method
func TestState_Merge_DeadlockPrevention(t *testing.T) {
	log := logrus.New()
	log.SetOutput(logrus.StandardLogger().Out)
	log.SetLevel(logrus.DebugLevel)

	// Create two states
	stateA := NewState(log)
	stateB := NewState(log)

	// Add test data
	stateA.UpdateSplit("source1", "hash1", "controller1")
	stateB.UpdateSplit("source2", "hash2", "controller2")

	// Create a wait group for concurrent merge operations
	var wg sync.WaitGroup
	wg.Add(2)

	// Run two merge operations concurrently in opposite directions
	// This would deadlock without the fix we implemented
	go func() {
		defer wg.Done()
		stateA.Merge(stateB)
	}()

	go func() {
		defer wg.Done()
		stateB.Merge(stateA)
	}()

	// Wait for both operations to complete
	// If there's a deadlock, the test will hang and eventually timeout
	wg.Wait()

	// Verify the merged results - both states should have all sources
	stateA.mu.Lock()
	assert.Contains(t, stateA.Sources, "source1", "StateA should contain source1")
	assert.Contains(t, stateA.Sources, "source2", "StateA should contain source2")
	assert.Equal(t, "hash1", stateA.Sources["source1"].Hash)
	assert.Equal(t, "controller1", stateA.Sources["source1"].Controller)
	stateA.mu.Unlock()

	stateB.mu.Lock()
	assert.Contains(t, stateB.Sources, "source1", "StateB should contain source1")
	assert.Contains(t, stateB.Sources, "source2", "StateB should contain source2")
	assert.Equal(t, "hash2", stateB.Sources["source2"].Hash)
	assert.Equal(t, "controller2", stateB.Sources["source2"].Controller)
	stateB.mu.Unlock()
}