package safe

import (
	"fmt"
	"time"
)

type RecoveryTransactionConfirmations struct {
	Timestamp             time.Time
	SafeAddress           string
	Group                 string
	Monitor               string
	RecoveryTransactionID string
	NumConfirmations      int
	ExpectedConfirmations int
}

const (
	RecoveryTransactionConfirmationsType = "safe_recovery_transaction_confirmations"
)

func NewRecoveryTransactionConfirmations(timestamp time.Time, monitor, group, safeAddress, recoveryTxID string, numConfirmations, expectedConfirmations int) *RecoveryTransactionConfirmations {
	return &RecoveryTransactionConfirmations{
		Timestamp:             timestamp,
		SafeAddress:           safeAddress,
		Group:                 group,
		Monitor:               monitor,
		RecoveryTransactionID: recoveryTxID,
		NumConfirmations:      numConfirmations,
		ExpectedConfirmations: expectedConfirmations,
	}
}

func (v *RecoveryTransactionConfirmations) GetType() string {
	return RecoveryTransactionConfirmationsType
}

func (v *RecoveryTransactionConfirmations) GetGroup() string {
	return v.Group
}

func (v *RecoveryTransactionConfirmations) GetMonitor() string {
	return v.Monitor
}

func (v *RecoveryTransactionConfirmations) GetTitle() string {
	return fmt.Sprintf("[%s] %s safe account has a recovery transaction with incorrect number of confirmations", v.Monitor, v.Group)
}

func (v *RecoveryTransactionConfirmations) GetDescription() string {
	return fmt.Sprintf(`
Timestamp: %s
Monitor: %s
Group: %s
Safe Account: %s (https://app.safe.global/home?safe=%s)
Recovery Transaction: %s (https://app.safe.global/transactions/queue?safe=%s)
Current Confirmations: %d
Expected Confirmations: %d`, v.Timestamp.UTC().Format("2006-01-02 15:04:05 UTC"), v.Monitor, v.Group, v.SafeAddress, v.SafeAddress, v.RecoveryTransactionID, v.SafeAddress, v.NumConfirmations, v.ExpectedConfirmations)
}
