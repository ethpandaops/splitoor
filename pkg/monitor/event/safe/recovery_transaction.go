package validator

import (
	"fmt"
	"time"
)

type RecoveryTransaction struct {
	Timestamp       time.Time
	SafeAddress     string
	Source          string
	Group           string
	Exists          bool
	NumSigners      int
	ExpectedSigners int
}

const (
	RecoveryTransactionType = "safe_recovery_transaction"
)

func (v *RecoveryTransaction) GetType() string {
	return RecoveryTransactionType
}

func (v *RecoveryTransaction) GetGroup() string {
	return v.Group
}
func (v *RecoveryTransaction) GetText() string {
	if !v.Exists {
		return fmt.Sprintf("Safe %s has no recovery transaction", v.SafeAddress)
	}
	return fmt.Sprintf("Safe %s has a recovery transaction with %d/%d signatures", v.SafeAddress, v.NumSigners, v.ExpectedSigners)
}

func (v *RecoveryTransaction) GetMarkdown() string {
	if !v.Exists {
		return fmt.Sprintf("Safe %s has no recovery transaction", v.SafeAddress)
	}
	return fmt.Sprintf("Safe %s has a recovery transaction with %d/%d signatures", v.SafeAddress, v.NumSigners, v.ExpectedSigners)
}
