package alert

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type Next struct {
	log logrus.FieldLogger

	alerting           bool
	hasRecoveryTx      bool
	hasRecoveryTxError bool
	recoveryTxIsNext   bool

	mu sync.Mutex
}

func NewNext(log logrus.FieldLogger) *Next {
	return &Next{
		log: log,
	}
}

func (n *Next) Update(hasRecoveryTx, hasRecoveryTxError, recoveryTxIsNext bool) (shouldAlert bool) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// only alert if the recovery tx is not next and there is a valid recovery tx
	shouldBeAlerting := !recoveryTxIsNext && hasRecoveryTx && !hasRecoveryTxError

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

	n.hasRecoveryTx = hasRecoveryTx
	n.hasRecoveryTxError = hasRecoveryTxError
	n.recoveryTxIsNext = recoveryTxIsNext

	return
}
