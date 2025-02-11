package alert

import (
	"github.com/sirupsen/logrus"
)

type Signers struct {
	log logrus.FieldLogger

	lastState bool
}

func NewSigners(log logrus.FieldLogger) *Signers {
	return &Signers{
		log: log.WithField("alert", "signers"),
	}
}

// Update returns true if an alert should be triggered
func (a *Signers) Update(mismatch bool) bool {
	defer func() {
		a.lastState = mismatch
	}()

	// Only alert on state change to true
	if !a.lastState && mismatch {
		return true
	}

	return false
}
