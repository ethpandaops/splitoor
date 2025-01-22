package safe

import (
	"fmt"
	"time"
)

type RecoveryTransactionMissing struct {
	Timestamp   time.Time
	SafeAddress string
	Group       string
	Monitor     string
}

const (
	RecoveryTransactionMissingType = "safe_recovery_transaction_missing"
)

func NewRecoveryTransactionMissing(timestamp time.Time, monitor, group, safeAddress string) *RecoveryTransactionMissing {
	return &RecoveryTransactionMissing{
		Timestamp:   timestamp,
		SafeAddress: safeAddress,
		Group:       group,
		Monitor:     monitor,
	}
}

func (v *RecoveryTransactionMissing) GetType() string {
	return RecoveryTransactionMissingType
}

func (v *RecoveryTransactionMissing) GetGroup() string {
	return v.Group
}

func (v *RecoveryTransactionMissing) GetMonitor() string {
	return v.Monitor
}

func (v *RecoveryTransactionMissing) GetTitle() string {
	return fmt.Sprintf("[%s] %s safe account has no recovery transaction queued", v.Monitor, v.Group)
}

func (v *RecoveryTransactionMissing) GetDescription() string {
	return fmt.Sprintf(`
Timestamp: %s
Monitor: %s
Group: %s
Safe Account: %s (https://app.safe.global/home?safe=%s)`, v.Timestamp.UTC().Format("2006-01-02 15:04:05 UTC"), v.Monitor, v.Group, v.SafeAddress, v.SafeAddress)
}
