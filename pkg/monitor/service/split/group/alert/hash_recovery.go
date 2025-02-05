package alert

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type HashRecovery struct {
	log          logrus.FieldLogger
	recoveryHash string

	alerting bool
	hash     string
	mu       sync.Mutex
}

func NewHashRecovery(log logrus.FieldLogger, recoveryHash string) *HashRecovery {
	return &HashRecovery{
		log:          log,
		recoveryHash: recoveryHash,
	}
}

func (b *HashRecovery) Update(hash string) (shouldAlert bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	shouldBeAlerting := hash == b.recoveryHash

	if b.alerting {
		shouldAlert = false

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
