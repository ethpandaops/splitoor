package alert

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type Confirmations struct {
	log logrus.FieldLogger

	alerting              bool
	numConfirmations      int
	expectedConfirmations int

	mu sync.Mutex
}

func NewConfirmations(log logrus.FieldLogger) *Confirmations {
	return &Confirmations{
		log: log,
	}
}

func (c *Confirmations) Update(numConfirmations, expectedConfirmations int) (shouldAlert bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	shouldBeAlerting := c.check(numConfirmations, expectedConfirmations)

	if c.alerting {
		shouldAlert = false

		if !shouldBeAlerting {
			c.alerting = false
		}
	} else {
		shouldAlert = false

		if shouldBeAlerting {
			c.alerting = true
			shouldAlert = true
		}
	}

	c.numConfirmations = numConfirmations
	c.expectedConfirmations = expectedConfirmations

	return
}

func (c *Confirmations) check(confirms, expected int) bool {
	return confirms != expected
}
