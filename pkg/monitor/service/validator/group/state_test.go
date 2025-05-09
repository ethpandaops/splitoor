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

	// Initialize validator maps
	stateA.Validators = make(map[string]*Validators)
	stateB.Validators = make(map[string]*Validators)

	// Add test data 
	stateA.Validators["pubkey1"] = &Validators{
		Sources: map[string]*Validator{
			"sourceA": {
				Balance:                    1000,
				Status:                     "active",
				WithdrawalCredentialsCode: 0,
			},
		},
	}

	stateB.Validators["pubkey2"] = &Validators{
		Sources: map[string]*Validator{
			"sourceB": {
				Balance:                    2000,
				Status:                     "pending",
				WithdrawalCredentialsCode: 1,
			},
		},
	}

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

	// Verify the merged results - both states should now have each other's data
	stateA.mu.Lock()
	assert.Contains(t, stateA.Validators, "pubkey1", "StateA should contain pubkey1")
	assert.Contains(t, stateA.Validators, "pubkey2", "StateA should contain pubkey2")
	assert.Contains(t, stateA.Validators["pubkey1"].Sources, "sourceA", "StateA should have sourceA for pubkey1")
	assert.Equal(t, "active", string(stateA.Validators["pubkey1"].Sources["sourceA"].Status))
	assert.Equal(t, uint64(1000), stateA.Validators["pubkey1"].Sources["sourceA"].Balance)
	stateA.mu.Unlock()

	stateB.mu.Lock()
	assert.Contains(t, stateB.Validators, "pubkey1", "StateB should contain pubkey1")
	assert.Contains(t, stateB.Validators, "pubkey2", "StateB should contain pubkey2")
	assert.Contains(t, stateB.Validators["pubkey2"].Sources, "sourceB", "StateB should have sourceB for pubkey2")
	assert.Equal(t, "pending", string(stateB.Validators["pubkey2"].Sources["sourceB"].Status))
	assert.Equal(t, uint64(2000), stateB.Validators["pubkey2"].Sources["sourceB"].Balance)
	stateB.mu.Unlock()
}