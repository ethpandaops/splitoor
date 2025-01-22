package alert

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type Invalid struct {
	log logrus.FieldLogger

	alerting bool
	invalid  error

	mu sync.Mutex
}

func NewInvalid(log logrus.FieldLogger) *Invalid {
	return &Invalid{
		log: log,
	}
}

func (m *Invalid) Update(invalid error) (shouldAlert bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	shouldBeAlerting := invalid != nil

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

	m.invalid = invalid

	return
}
