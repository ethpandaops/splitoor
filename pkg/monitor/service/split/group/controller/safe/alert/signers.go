package alert

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type Signers struct {
	log logrus.FieldLogger

	alerting  bool
	lastState bool

	mu sync.Mutex
}

func NewSigners(log logrus.FieldLogger) *Signers {
	return &Signers{
		log: log.WithField("alert", "signers"),
	}
}

// Update returns true if an alert should be triggered
func (a *Signers) Update(mismatch bool) (shouldAlert bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	shouldBeAlerting := mismatch

	if a.alerting {
		shouldAlert = false

		if !shouldBeAlerting {
			a.alerting = false
		}
	} else {
		shouldAlert = false

		if shouldBeAlerting {
			a.alerting = true
			shouldAlert = true
		}
	}

	a.lastState = mismatch

	return
}
