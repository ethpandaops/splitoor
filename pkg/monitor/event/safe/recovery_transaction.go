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
	Monitor         string
	TxnHash         string
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

func (v *RecoveryTransaction) GetMonitor() string {
	return v.Monitor
}

func (v *RecoveryTransaction) GetTitle() string {
	if v.TxnHash == "" {
		return fmt.Sprintf("[%s] %s safe account has no recovery transaction", v.Monitor, v.Group)
	}

	return fmt.Sprintf("[%s] %s safe account has a recovery transaction with incorrect number of signatures", v.Monitor, v.Group)
}

func (v *RecoveryTransaction) GetDescription() string {
	if v.TxnHash == "" {
		return fmt.Sprintf(`
Timestamp: %s
Monitor: %s
Group: %s
Source: %s
Safe Account: %s (https://app.safe.global/home?safe=%s)`, v.Timestamp.UTC().Format("2006-01-02 15:04:05 UTC"), v.Monitor, v.Group, v.Source, v.SafeAddress, v.SafeAddress)
	}

	return fmt.Sprintf(`
Timestamp: %s
Monitor: %s
Group: %s
Source: %s
Safe Account: %s (https://app.safe.global/home?safe=%s)
Recovery Transaction: %s (https://app.safe.global/transactions/queue?safe=%s)
Signatures: %d/%d`, v.Timestamp.UTC().Format("2006-01-02 15:04:05 UTC"), v.Monitor, v.Group, v.Source, v.SafeAddress, v.SafeAddress, v.TxnHash, v.SafeAddress, v.NumSigners, v.ExpectedSigners)
}
