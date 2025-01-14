package validator

import (
	"fmt"
	"time"
)

type TransactionQueue struct {
	Timestamp   time.Time
	SafeAddress string
	Source      string
	Group       string
	Monitor     string
	NumTxs      int
}

const (
	TransactionQueueType = "safe_transaction_queue"
)

func (v *TransactionQueue) GetType() string {
	return TransactionQueueType
}

func (v *TransactionQueue) GetGroup() string {
	return v.Group
}

func (v *TransactionQueue) GetMonitor() string {
	return v.Monitor
}

func (v *TransactionQueue) GetTitle() string {
	return fmt.Sprintf("[%s] %s safe has unexpected transactions", v.Monitor, v.Group)
}

func (v *TransactionQueue) GetDescription() string {
	return fmt.Sprintf(`
Timestamp: %s
Monitor: %s
Group: %s
Source: %s
Safe Account: %s (https://app.safe.global/home?safe=%s)
Number of Transactions: %d (https://app.safe.global/transactions/queue?safe=%s)`, v.Timestamp.UTC().Format("2006-01-02 15:04:05 UTC"), v.Monitor, v.Group, v.Source, v.SafeAddress, v.SafeAddress, v.NumTxs, v.SafeAddress)
}
