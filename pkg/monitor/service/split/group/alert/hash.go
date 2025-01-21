package alert

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type Hash struct {
	log          logrus.FieldLogger
	expectedHash string

	alerting bool
	hash     string
	mu       sync.Mutex
}

func NewHash(log logrus.FieldLogger, expectedHash string) *Hash {
	return &Hash{
		log:          log,
		expectedHash: expectedHash,
	}
}

func (b *Hash) Update(hash string) (shouldAlert bool, alertingHash *string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// check if any new balances trigger the alert
	shouldBeAlerting, alertingHash := b.check(hash)

	// if already alerting, check if should still be alerting
	if b.alerting {
		// shouldn't re-alert if already alerting
		shouldAlert = false
		// stop alerting if no longer should be alerting
		if !shouldBeAlerting {
			b.alerting = false
		}
	} else {
		shouldAlert = false

		if shouldBeAlerting {
			b.alerting = true
			shouldAlert = true
		}
	}

	b.hash = hash

	return
}

func (b *Hash) check(h string) (shouldAlert bool, hash *string) {
	if h != b.expectedHash {
		return true, &h
	}

	return false, nil
}
