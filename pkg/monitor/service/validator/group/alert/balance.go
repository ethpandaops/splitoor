package alert

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type Balance struct {
	log        logrus.FieldLogger
	minBalance uint64

	alerting bool
	balances []uint64
	mu       sync.Mutex
}

func NewBalance(log logrus.FieldLogger, minBalance uint64) *Balance {
	return &Balance{
		log:        log,
		minBalance: minBalance,
	}
}

func (b *Balance) Update(balances []uint64) (shouldAlert bool, balance *uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// check if any new balances trigger the alert
	shouldBeAlerting, balance := b.check(balances)

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

	b.balances = balances

	return
}

func (b *Balance) check(balances []uint64) (shouldAlert bool, balance *uint64) {
	for _, balance := range balances {
		if balance < b.minBalance {
			return true, &balance
		}
	}

	return false, nil
}
