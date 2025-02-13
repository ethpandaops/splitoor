package safe

import (
	"fmt"
	"strings"
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

func (v *TransactionQueueExcess) GetTitle(includeMonitor, includeGroup bool) string {
	var sb strings.Builder

	if includeMonitor {
		sb.WriteString("[")
		sb.WriteString(v.Monitor)
		sb.WriteString("] ")
	}

	sb.WriteString("Safe has unexpected transactions in queue")

	return sb.String()
}

func (v *TransactionQueueExcess) GetDescriptionText(includeMonitor, includeGroup bool) string {
	var sb strings.Builder

	sb.WriteString("\nTimestamp: ")
	sb.WriteString(v.Timestamp.UTC().Format("2006-01-02 15:04:05 UTC"))

	if includeMonitor {
		sb.WriteString("\nMonitor: ")
		sb.WriteString(v.Monitor)
	}

	if includeGroup {
		sb.WriteString("\nGroup: ")
		sb.WriteString(v.Group)
	}

	sb.WriteString("\nSafe Account: ")
	sb.WriteString(v.SafeAddress)

	sb.WriteString("\nNumber of Transactions: ")
	sb.WriteString(fmt.Sprintf("%d", v.NumTxs))

	return sb.String()
}

func (v *TransactionQueueExcess) GetDescriptionMarkdown(includeMonitor, includeGroup bool) string {
	var sb strings.Builder

	sb.WriteString("**Timestamp:** ")
	sb.WriteString(v.Timestamp.UTC().Format("2006-01-02 15:04:05 UTC"))
	sb.WriteString("\n")

	if includeMonitor {
		sb.WriteString("**Monitor:** ")
		sb.WriteString(v.Monitor)
		sb.WriteString("\n")
	}

	if includeGroup {
		sb.WriteString("**Group:** ")
		sb.WriteString(v.Group)
		sb.WriteString("\n")
	}

	sb.WriteString("**Safe Account:** `")
	sb.WriteString(v.SafeAddress)
	sb.WriteString("`\n")

	sb.WriteString("**Number of Transactions:** ")
	sb.WriteString(fmt.Sprintf("%d", v.NumTxs))

	return sb.String()
}

func (v *TransactionQueueExcess) GetDescriptionHTML(includeMonitor, includeGroup bool) string {
	var sb strings.Builder

	sb.WriteString("<p><strong>Timestamp:</strong> ")
	sb.WriteString(v.Timestamp.UTC().Format("2006-01-02 15:04:05 UTC"))
	sb.WriteString("</p>")

	if includeMonitor {
		sb.WriteString("<p><strong>Monitor:</strong> ")
		sb.WriteString(v.Monitor)
		sb.WriteString("</p>")
	}

	if includeGroup {
		sb.WriteString("<p><strong>Group:</strong> ")
		sb.WriteString(v.Group)
		sb.WriteString("</p>")
	}

	sb.WriteString("<p><strong>Safe Account:</strong> ")
	sb.WriteString(v.SafeAddress)
	sb.WriteString("</p>")

	sb.WriteString("<p><strong>Number of Transactions:</strong> ")
	sb.WriteString(fmt.Sprintf("%d", v.NumTxs))
	sb.WriteString("</p>")

	return sb.String()
}
