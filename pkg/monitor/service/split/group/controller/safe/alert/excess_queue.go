package alert

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type ExcessQueue struct {
	log logrus.FieldLogger

	alerting bool
	length   int
	maxLen   int

	mu sync.Mutex
}

func NewExcessQueue(log logrus.FieldLogger, maxLen int) *ExcessQueue {
	return &ExcessQueue{
		log:    log,
		maxLen: maxLen,
	}
}

func (e *ExcessQueue) Update(length int) (shouldAlert bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	shouldBeAlerting := e.check(length)

	if e.alerting {
		shouldAlert = false

		if !shouldBeAlerting {
			e.alerting = false
		}
	} else {
		shouldAlert = false

		if shouldBeAlerting {
			e.alerting = true
			shouldAlert = true
		}
	}

	e.length = length

	return
}

func (e *ExcessQueue) check(length int) bool {
	return length > e.maxLen
}

func (e *ExcessQueue) Alerting() bool {
	return e.alerting
}
