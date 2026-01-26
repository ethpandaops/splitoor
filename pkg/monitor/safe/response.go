package safe

import "strings"

// MultisigTransactionsResponse represents the response from /multisig-transactions/?executed=false.
type MultisigTransactionsResponse struct {
	Count            int                   `json:"count"`
	Next             *string               `json:"next"`
	Previous         *string               `json:"previous"`
	Results          []MultisigTransaction `json:"results"`
	CountUniqueNonce int                   `json:"countUniqueNonce"`
}

// MultisigTransaction represents a multisig transaction from the Transaction Service API v2.
// Used for both list and detail endpoints.
type MultisigTransaction struct {
	Safe                  string         `json:"safe"`
	To                    string         `json:"to"`
	Value                 string         `json:"value"`
	Data                  *string        `json:"data"`
	Operation             int            `json:"operation"`
	SafeTxHash            string         `json:"safeTxHash"`
	Nonce                 int64          `json:"nonce,string"`
	SubmissionDate        string         `json:"submissionDate"`
	IsExecuted            bool           `json:"isExecuted"`
	ConfirmationsRequired int            `json:"confirmationsRequired"`
	Confirmations         []Confirmation `json:"confirmations"`
	DataDecoded           *DataDecoded   `json:"dataDecoded"`
	Proposer              string         `json:"proposer"`
	Origin                string         `json:"origin"`
	Trusted               bool           `json:"trusted"`
}

// Confirmation represents a transaction confirmation.
type Confirmation struct {
	Owner          string `json:"owner"`
	SubmissionDate string `json:"submissionDate"`
	Signature      string `json:"signature"`
	SignatureType  string `json:"signatureType"`
}

// DataDecoded represents decoded transaction data.
type DataDecoded struct {
	Method     string      `json:"method"`
	Parameters []Parameter `json:"parameters"`
}

// Parameter represents a decoded parameter from transaction data.
type Parameter struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Value        any    `json:"value"`
	ValueDecoded any    `json:"valueDecoded"`
}

// SafeResponse represents the response from /safes/{address}/.
type SafeResponse struct {
	Address         string   `json:"address"`
	Nonce           int64    `json:"nonce,string"`
	Threshold       int      `json:"threshold"`
	Owners          []string `json:"owners"`
	MasterCopy      string   `json:"masterCopy"`
	Modules         []string `json:"modules"`
	FallbackHandler string   `json:"fallbackHandler"`
	Guard           string   `json:"guard"`
	Version         *string  `json:"version"`
}

// CheckSigners verifies that the provided signers match the Safe owners.
func (s *SafeResponse) CheckSigners(signers []string) bool {
	if s == nil {
		return false
	}

	if len(signers) == 0 {
		return false
	}

	if len(s.Owners) == 0 {
		return false
	}

	// Build map of actual signers
	actualSigners := make(map[string]bool, len(s.Owners))
	for _, owner := range s.Owners {
		actualSigners[strings.ToLower(owner)] = true
	}

	// Check if all provided signers are actual signers
	for _, signer := range signers {
		if !actualSigners[strings.ToLower(signer)] {
			return false
		}
	}

	// Check if number of signers matches number of owners
	return len(signers) == len(s.Owners)
}
