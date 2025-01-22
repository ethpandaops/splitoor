package safe

import (
	"fmt"
	"time"
)

type RecoveryTransactionInvalid struct {
	Timestamp             time.Time
	SafeAddress           string
	Group                 string
	Monitor               string
	RecoveryTransactionID string
	Reason                string
}

const (
	RecoveryTransactionInvalidType = "safe_recovery_transaction_invalid"
)

func NewRecoveryTransactionInvalid(timestamp time.Time, monitor, group, safeAddress, txID, reason string) *RecoveryTransactionInvalid {
	return &RecoveryTransactionInvalid{
		Timestamp:             timestamp,
		SafeAddress:           safeAddress,
		Group:                 group,
		Monitor:               monitor,
		RecoveryTransactionID: txID,
		Reason:                reason,
	}
}

func (v *RecoveryTransactionInvalid) GetType() string {
	return RecoveryTransactionInvalidType
}

func (v *RecoveryTransactionInvalid) GetGroup() string {
	return v.Group
}

func (v *RecoveryTransactionInvalid) GetMonitor() string {
	return v.Monitor
}

func (v *RecoveryTransactionInvalid) GetTitle() string {
	return fmt.Sprintf("[%s] %s safe account has invalid recovery transaction", v.Monitor, v.Group)
}

func (v *RecoveryTransactionInvalid) GetDescription() string {
	return fmt.Sprintf(`
Timestamp: %s
Monitor: %s
Group: %s
Safe Account: %s (https://app.safe.global/home?safe=%s)
Transaction ID: %s
Reason: %s`, v.Timestamp.UTC().Format("2006-01-02 15:04:05 UTC"), v.Monitor, v.Group, v.SafeAddress, v.SafeAddress, v.RecoveryTransactionID, v.Reason)
}
