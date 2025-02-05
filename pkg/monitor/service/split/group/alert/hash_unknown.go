package alert

import (
	"slices"
	"sync"

	"github.com/sirupsen/logrus"
)

type HashUnknown struct {
	log            logrus.FieldLogger
	expectedHashes []string

	alerting bool
	hash     string
	mu       sync.Mutex
}

func NewHashUnknown(log logrus.FieldLogger, expectedHashes []string) *HashUnknown {
	return &HashUnknown{
		log:            log,
		expectedHashes: expectedHashes,
	}
}

func (b *HashUnknown) Update(hash string) (shouldAlert bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	shouldBeAlerting := !slices.Contains(b.expectedHashes, hash)

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
