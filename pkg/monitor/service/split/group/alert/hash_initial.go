package alert

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type HashInitial struct {
	log         logrus.FieldLogger
	initialHash string

	alerting bool
	hash     string
	mu       sync.Mutex
}

func NewHashInitial(log logrus.FieldLogger, initialHash string) *HashInitial {
	return &HashInitial{
		log:         log,
		initialHash: initialHash,
	}
}

func (b *HashInitial) Update(hash string) (shouldAlert bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	shouldBeAlerting := hash == b.initialHash

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
