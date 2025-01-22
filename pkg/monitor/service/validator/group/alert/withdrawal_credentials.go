package alert

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type WithdrawalCredentials struct {
	log        logrus.FieldLogger
	allowedSet map[int64]struct{}

	alerting bool
	codes    []int64
	mu       sync.Mutex
}

func NewWithdrawalCredentials(log logrus.FieldLogger, allowed []int64) *WithdrawalCredentials {
	allowedSet := make(map[int64]struct{}, len(allowed))
	for _, a := range allowed {
		allowedSet[a] = struct{}{}
	}

	return &WithdrawalCredentials{
		log:        log,
		allowedSet: allowedSet,
	}
}

func (w *WithdrawalCredentials) Update(codes []int64) (shouldAlert bool, alertingCredential *int64) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// check if any new credentials trigger the alert
	shouldBeAlerting, alertingCredential := w.check(codes)

	// if already alerting, check if should still be alerting
	if w.alerting {
		// shouldn't re-alert if already alerting
		shouldAlert = false
		// stop alerting if no longer should be alerting
		if !shouldBeAlerting {
			w.alerting = false
		}
	} else {
		shouldAlert = false

		if shouldBeAlerting {
			w.alerting = true
			shouldAlert = true
		}
	}

	w.codes = codes

	return
}

func (w *WithdrawalCredentials) check(codes []int64) (shouldAlert bool, alertingCredential *int64) {
	for _, code := range codes {
		if _, exists := w.allowedSet[code]; !exists {
			return true, &code
		}
	}

	return false, nil
}
