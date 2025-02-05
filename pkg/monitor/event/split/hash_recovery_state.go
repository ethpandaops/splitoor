package split

import (
	"strings"
	"time"
)

type HashRecoveryState struct {
	Timestamp    time.Time
	SplitAddress string
	Hash         string
	Group        string
	Monitor      string
}

const (
	HashRecoveryStateType = "split_hash_recovery_state"
)

func NewHashRecoveryState(timestamp time.Time, monitor, group, splitAddress, hash string) *HashRecoveryState {
	return &HashRecoveryState{
		Timestamp:    timestamp,
		SplitAddress: splitAddress,
		Hash:         hash,
		Group:        group,
		Monitor:      monitor,
	}
}

func (v *HashRecoveryState) GetType() string {
	return HashRecoveryStateType
}

func (v *HashRecoveryState) GetGroup() string {
	return v.Group
}

func (v *HashRecoveryState) GetMonitor() string {
	return v.Monitor
}

func (v *HashRecoveryState) GetTitle(includeMonitor, includeGroup bool) string {
	var sb strings.Builder

	if includeMonitor {
		sb.WriteString("[")
		sb.WriteString(v.Monitor)
		sb.WriteString("] ")
	}

	sb.WriteString("Split hash is in recovery state")

	return sb.String()
}

func (v *HashRecoveryState) GetDescriptionText(includeMonitor, includeGroup bool) string {
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

	sb.WriteString("\nSplit Address: ")
	sb.WriteString(v.SplitAddress)
	sb.WriteString("\nHash: ")
	sb.WriteString(v.Hash)

	return sb.String()
}

func (v *HashRecoveryState) GetDescriptionMarkdown(includeMonitor, includeGroup bool) string {
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

	sb.WriteString("**Split Address:** `")
	sb.WriteString(v.SplitAddress)
	sb.WriteString("`\n")

	sb.WriteString("**Hash:** `")
	sb.WriteString(v.Hash)
	sb.WriteString("`")

	return sb.String()
}

func (v *HashRecoveryState) GetDescriptionHTML(includeMonitor, includeGroup bool) string {
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

	sb.WriteString("<p><strong>Split Address:</strong> ")
	sb.WriteString(v.SplitAddress)
	sb.WriteString("</p>")

	sb.WriteString("<p><strong>Hash:</strong> ")
	sb.WriteString(v.Hash)
	sb.WriteString("</p>")

	return sb.String()
}
