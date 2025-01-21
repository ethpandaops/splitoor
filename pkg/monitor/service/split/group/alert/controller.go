package alert

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type Controller struct {
	log                logrus.FieldLogger
	expectedController string

	alerting   bool
	controller string
	mu         sync.Mutex
}

func NewController(log logrus.FieldLogger, expectedController string) *Controller {
	return &Controller{
		log:                log,
		expectedController: expectedController,
	}
}

func (b *Controller) Update(controller string) (shouldAlert bool, alertingController *string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// check if any new balances trigger the alert
	shouldBeAlerting, alertingController := b.check(controller)

	// if already alerting, check if should still be alerting
	if b.alerting {
		// shouldn't re-alert if already alerting
		shouldAlert = false
		// stop alerting if no longer should be alerting
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

	b.controller = controller

	return
}

func (b *Controller) check(c string) (shouldAlert bool, controller *string) {
	if c != b.expectedController {
		return true, &c
	}

	return false, nil
}
