package safe

import (
	"fmt"
	"time"
)

type RecoveryTransactionNotNext struct {
	Timestamp             time.Time
	SafeAddress           string
	Group                 string
	Monitor               string
	RecoveryTransactionID string
}

const (
	RecoveryTransactionNotNextType = "safe_recovery_transaction_not_next"
)

func NewRecoveryTransactionNotNext(timestamp time.Time, monitor, group, safeAddress, recoveryTxID string) *RecoveryTransactionNotNext {
	return &RecoveryTransactionNotNext{
		Timestamp:             timestamp,
		SafeAddress:           safeAddress,
		Group:                 group,
		Monitor:               monitor,
		RecoveryTransactionID: recoveryTxID,
	}
}

func (v *RecoveryTransactionNotNext) GetType() string {
	return RecoveryTransactionNotNextType
}

func (v *RecoveryTransactionNotNext) GetGroup() string {
	return v.Group
}

func (v *RecoveryTransactionNotNext) GetMonitor() string {
	return v.Monitor
}

func (v *RecoveryTransactionNotNext) GetTitle() string {
	return fmt.Sprintf("[%s] %s safe account has a recovery transaction that is not next in queue", v.Monitor, v.Group)
}

func (v *RecoveryTransactionNotNext) GetDescription() string {
	return fmt.Sprintf(`
Timestamp: %s
Monitor: %s
Group: %s
Safe Account: %s (https://app.safe.global/home?safe=%s)
Recovery Transaction: %s (https://app.safe.global/transactions/queue?safe=%s)`, v.Timestamp.UTC().Format("2006-01-02 15:04:05 UTC"), v.Monitor, v.Group, v.SafeAddress, v.SafeAddress, v.RecoveryTransactionID, v.SafeAddress)
}
