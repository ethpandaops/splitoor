package validator

import (
	"fmt"
	"strings"
	"time"
)

type MinBalance struct {
	Timestamp time.Time
	Pubkey    string
	Balance   uint64
	Group     string
	Monitor   string
}

const (
	MinBalanceType = "validator_min_balance"
)

func NewMinBalance(timestamp time.Time, balance uint64, pubkey, group, monitor string) *MinBalance {
	return &MinBalance{
		Timestamp: timestamp,
		Pubkey:    pubkey,
		Balance:   balance,
		Group:     group,
		Monitor:   monitor,
	}
}

func (v *MinBalance) GetType() string {
	return MinBalanceType
}

func (v *MinBalance) GetGroup() string {
	return v.Group
}

func (v *MinBalance) GetMonitor() string {
	return v.Monitor
}

func (v *MinBalance) GetTitle(includeMonitor, includeGroup bool) string {
	var sb strings.Builder

	if includeMonitor {
		sb.WriteString("[")
		sb.WriteString(v.Monitor)
		sb.WriteString("] ")
	}

	sb.WriteString("Validator has low balance")

	return sb.String()
}

func (v *MinBalance) GetDescriptionText(includeMonitor, includeGroup bool) string {
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

	sb.WriteString("\nPubkey: ")
	sb.WriteString(v.Pubkey)
	sb.WriteString("\nBalance: ")
	sb.WriteString(fmt.Sprintf("%.4f ETH", float64(v.Balance)/1e18))

	return sb.String()
}

func (v *MinBalance) GetDescriptionMarkdown(includeMonitor, includeGroup bool) string {
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

	sb.WriteString("**Pubkey:** `")
	sb.WriteString(v.Pubkey)
	sb.WriteString("`\n")

	sb.WriteString("**Balance:** ")
	sb.WriteString(fmt.Sprintf("%.4f ETH", float64(v.Balance)/1e18))

	return sb.String()
}

func (v *MinBalance) GetDescriptionHTML(includeMonitor, includeGroup bool) string {
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

	sb.WriteString("<p><strong>Pubkey:</strong> ")
	sb.WriteString(v.Pubkey)
	sb.WriteString("</p>")

	sb.WriteString("<p><strong>Balance:</strong> ")
	sb.WriteString(fmt.Sprintf("%.4f ETH", float64(v.Balance)/1e18))
	sb.WriteString("</p>")

	return sb.String()
}
