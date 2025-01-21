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

func (b *Controller) Update(controller string) (shouldAlert bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	shouldBeAlerting := controller != b.expectedController

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

	b.controller = controller

	return
}
