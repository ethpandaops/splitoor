package split

import (
	"strings"
	"time"
)

type HashUnknownState struct {
	Timestamp    time.Time
	SplitAddress string
	ExpectedHash string
	ActualHash   string
	Group        string
	Monitor      string
}

const (
	HashUnknownStateType = "split_hash_unknown_state"
)

func NewHashUnknownState(timestamp time.Time, monitor, group, splitAddress, expectedHash, actualHash string) *HashUnknownState {
	return &HashUnknownState{
		Timestamp:    timestamp,
		SplitAddress: splitAddress,
		ExpectedHash: expectedHash,
		ActualHash:   actualHash,
		Group:        group,
		Monitor:      monitor,
	}
}

func (v *HashUnknownState) GetType() string {
	return HashUnknownStateType
}

func (v *HashUnknownState) GetGroup() string {
	return v.Group
}

func (v *HashUnknownState) GetMonitor() string {
	return v.Monitor
}

func (v *HashUnknownState) GetTitle(includeMonitor, includeGroup bool) string {
	var sb strings.Builder

	if includeMonitor {
		sb.WriteString("[")
		sb.WriteString(v.Monitor)
		sb.WriteString("] ")
	}

	sb.WriteString("Split hash is in unknown state")

	return sb.String()
}

func (v *HashUnknownState) GetDescriptionText(includeMonitor, includeGroup bool) string {
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
	sb.WriteString("\nExpected Hash: ")
	sb.WriteString(v.ExpectedHash)
	sb.WriteString("\nActual Hash: ")
	sb.WriteString(v.ActualHash)

	return sb.String()
}

func (v *HashUnknownState) GetDescriptionMarkdown(includeMonitor, includeGroup bool) string {
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

	sb.WriteString("**Expected Hash:** `")
	sb.WriteString(v.ExpectedHash)
	sb.WriteString("`\n")

	sb.WriteString("**Actual Hash:** `")
	sb.WriteString(v.ActualHash)
	sb.WriteString("`")

	return sb.String()
}

func (v *HashUnknownState) GetDescriptionHTML(includeMonitor, includeGroup bool) string {
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

	sb.WriteString("<p><strong>Expected Hash:</strong> ")
	sb.WriteString(v.ExpectedHash)
	sb.WriteString("</p>")

	sb.WriteString("<p><strong>Actual Hash:</strong> ")
	sb.WriteString(v.ActualHash)
	sb.WriteString("</p>")

	return sb.String()
}
