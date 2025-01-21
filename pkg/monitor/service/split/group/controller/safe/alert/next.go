package alert

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type Next struct {
	log logrus.FieldLogger

	alerting bool
	isNext   bool

	mu sync.Mutex
}

func NewNext(log logrus.FieldLogger) *Next {
	return &Next{
		log: log,
	}
}

func (n *Next) Update(isNext bool) (shouldAlert bool) {
	n.mu.Lock()
	defer n.mu.Unlock()

	shouldBeAlerting := !isNext

	if n.alerting {
		shouldAlert = false

		if !shouldBeAlerting {
			n.alerting = false
		}
	} else {
		shouldAlert = false

		if shouldBeAlerting {
			n.alerting = true
			shouldAlert = true
		}
	}

	n.isNext = isNext

	return
}
