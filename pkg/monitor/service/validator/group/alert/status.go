package alert

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type Status struct {
	log        logrus.FieldLogger
	allowedSet map[string]struct{}

	alerting bool
	statuses []string
	mu       sync.Mutex
}

func NewStatus(log logrus.FieldLogger, allowed []string) *Status {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, a := range allowed {
		allowedSet[a] = struct{}{}
	}

	return &Status{
		log:        log,
		allowedSet: allowedSet,
	}
}

func (s *Status) Update(statuses []string) (shouldAlert bool, alertingStatus *string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// check if any new statuses trigger the alert
	shouldBeAlerting, alertingStatus := s.check(statuses)

	// if already alerting, check if should still be alerting
	if s.alerting {
		// shouldn't re-alert if already alerting
		shouldAlert = false
		// stop alerting if no longer should be alerting
		if !shouldBeAlerting {
			s.alerting = false
		}
	} else {
		shouldAlert = false

		if shouldBeAlerting {
			s.alerting = true
			shouldAlert = true
		}
	}

	s.statuses = statuses

	return
}

func (s *Status) check(statuses []string) (shouldAlert bool, alertingStatus *string) {
	for _, st := range statuses {
		if _, exists := s.allowedSet[st]; !exists {
			return true, &st
		}
	}

	return false, nil
}
