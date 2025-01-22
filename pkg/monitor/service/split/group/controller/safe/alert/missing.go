package alert

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type Missing struct {
	log logrus.FieldLogger

	alerting bool
	missing  bool

	mu sync.Mutex
}

func NewMissing(log logrus.FieldLogger) *Missing {
	return &Missing{
		log: log,
	}
}

func (m *Missing) Update(missing bool) (shouldAlert bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	shouldBeAlerting := missing

	if m.alerting {
		shouldAlert = false

		if !shouldBeAlerting {
			m.alerting = false
		}
	} else {
		shouldAlert = false

		if shouldBeAlerting {
			m.alerting = true
			shouldAlert = true
		}
	}

	m.missing = missing

	return
}
