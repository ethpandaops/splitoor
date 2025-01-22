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

func (b *Hash) Update(hash string) (shouldAlert bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	shouldBeAlerting := hash != b.expectedHash

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
