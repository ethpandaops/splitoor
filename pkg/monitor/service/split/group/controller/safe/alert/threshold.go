package alert

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type Threshold struct {
	log logrus.FieldLogger

	alerting          bool
	threshold         int
	expectedThreshold int

	mu sync.Mutex
}

func NewThreshold(log logrus.FieldLogger, expectedThreshold int) *Threshold {
	return &Threshold{
		log:               log.WithField("alert", "threshold"),
		expectedThreshold: expectedThreshold,
	}
}

// Update returns true if an alert should be triggered
func (t *Threshold) Update(threshold int) (shouldAlert bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	shouldBeAlerting := threshold != t.expectedThreshold

	if t.alerting {
		shouldAlert = false

		if !shouldBeAlerting {
			t.alerting = false
		}
	} else {
		shouldAlert = false

		if shouldBeAlerting {
			t.alerting = true
			shouldAlert = true
		}
	}

	t.threshold = threshold

	return
}
