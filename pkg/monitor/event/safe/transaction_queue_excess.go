package safe

import (
	"fmt"
	"time"
)

type TransactionQueueExcess struct {
	Timestamp   time.Time
	SafeAddress string
	Group       string
	Monitor     string
	NumTxs      int
}

const (
	TransactionQueueExcessType = "safe_transaction_queue_excess"
)

func NewTransactionQueueExcess(timestamp time.Time, monitor, group, safeAddress string, numTxs int) *TransactionQueueExcess {
	return &TransactionQueueExcess{
		Timestamp:   timestamp,
		SafeAddress: safeAddress,
		Group:       group,
		Monitor:     monitor,
		NumTxs:      numTxs,
	}
}

func (v *TransactionQueueExcess) GetType() string {
	return TransactionQueueExcessType
}

func (v *TransactionQueueExcess) GetGroup() string {
	return v.Group
}

func (v *TransactionQueueExcess) GetMonitor() string {
	return v.Monitor
}

func (v *TransactionQueueExcess) GetTitle() string {
	return fmt.Sprintf("[%s] %s safe has unexpected transactions in queue", v.Monitor, v.Group)
}

func (v *TransactionQueueExcess) GetDescription() string {
	return fmt.Sprintf(`
Timestamp: %s
Monitor: %s
Group: %s
Safe Account: %s (https://app.safe.global/home?safe=%s)
Number of Transactions: %d (https://app.safe.global/transactions/queue?safe=%s)`, v.Timestamp.UTC().Format("2006-01-02 15:04:05 UTC"), v.Monitor, v.Group, v.SafeAddress, v.SafeAddress, v.NumTxs, v.SafeAddress)
}
