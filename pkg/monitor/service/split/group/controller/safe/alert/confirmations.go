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

func (c *Confirmations) Update(numConfirmations, expectedConfirmations int, hasNextRecoveryTx bool) (shouldAlert bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// only alert if the number of confirmations is not the expected number and there is a valid recovery tx that is next in the queue
	shouldBeAlerting := numConfirmations != expectedConfirmations && hasNextRecoveryTx

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
