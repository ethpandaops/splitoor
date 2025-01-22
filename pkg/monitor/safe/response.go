package safe

type QueuedTransactionsResponse struct {
	Count    int                       `json:"count"`
	Next     *string                   `json:"next"`
	Previous *string                   `json:"previous"`
	Results  []QueuedTransactionResult `json:"results"`
}

type QueuedTransactionResult struct {
	Type         string       `json:"type"`
	Label        *string      `json:"label,omitempty"`
	Transaction  *Transaction `json:"transaction,omitempty"`
	ConflictType *string      `json:"conflictType,omitempty"`
}

type Transaction struct {
	ID            string          `json:"id"`
	Timestamp     int64           `json:"timestamp"`
	TxStatus      string          `json:"txStatus"`
	TxInfo        TransactionInfo `json:"txInfo"`
	ExecutionInfo ExecutionInfo   `json:"executionInfo"`
	SafeAppInfo   *SafeAppInfo    `json:"safeAppInfo"`
	TxHash        *string         `json:"txHash"`
}

type TransactionInfo struct {
	Type             string        `json:"type"`
	HumanDescription *string       `json:"humanDescription"`
	Sender           AddressInfo   `json:"sender"`
	Recipient        AddressInfo   `json:"recipient"`
	Direction        string        `json:"direction"`
	TransferInfo     *TransferInfo `json:"transferInfo,omitempty"`
}

type ExecutionInfo struct {
	Type                   string        `json:"type"`
	Nonce                  int           `json:"nonce"`
	ConfirmationsRequired  int           `json:"confirmationsRequired"`
	ConfirmationsSubmitted int           `json:"confirmationsSubmitted"`
	MissingSigners         []AddressInfo `json:"missingSigners,omitempty"`
}

type AddressInfo struct {
	Value   string  `json:"value"`
	Name    *string `json:"name"`
	LogoURI *string `json:"logoUri"`
}

type TransferInfo struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type SafeAppInfo struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	LogoURI string `json:"logoUri"`
}

type TransactionDetails struct {
	SafeAddress           string                `json:"safeAddress"`
	TxID                  string                `json:"txId"`
	ExecutedAt            *string               `json:"executedAt"`
	TxStatus              string                `json:"txStatus"`
	TxInfo                TransactionInfo       `json:"txInfo"`
	TxData                TransactionData       `json:"txData"`
	TxHash                *string               `json:"txHash"`
	DetailedExecutionInfo DetailedExecutionInfo `json:"detailedExecutionInfo"`
	SafeAppInfo           *SafeAppInfo          `json:"safeAppInfo"`
	Note                  *string               `json:"note"`
}

type TransactionData struct {
	HexData                   *string                 `json:"hexData"`
	DataDecoded               *DataDecoded            `json:"dataDecoded"`
	To                        AddressInfo             `json:"to"`
	Value                     string                  `json:"value"`
	Operation                 int                     `json:"operation"`
	TrustedDelegateCallTarget *string                 `json:"trustedDelegateCallTarget"`
	AddressInfoIndex          *map[string]AddressInfo `json:"addressInfoIndex"`
}

type DataDecoded struct {
	Method     string      `json:"method"`
	Parameters []Parameter `json:"parameters"`
}

type Parameter struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Value        interface{} `json:"value"`
	ValueDecoded interface{} `json:"valueDecoded"`
}

type DetailedExecutionInfo struct {
	Type                  string         `json:"type"`
	SubmittedAt           int64          `json:"submittedAt"`
	Nonce                 int            `json:"nonce"`
	SafeTxGas             string         `json:"safeTxGas"`
	BaseGas               string         `json:"baseGas"`
	GasPrice              string         `json:"gasPrice"`
	GasToken              string         `json:"gasToken"`
	RefundReceiver        AddressInfo    `json:"refundReceiver"`
	SafeTxHash            string         `json:"safeTxHash"`
	Executor              *AddressInfo   `json:"executor"`
	Signers               []AddressInfo  `json:"signers"`
	ConfirmationsRequired int            `json:"confirmationsRequired"`
	Confirmations         []Confirmation `json:"confirmations"`
	Rejectors             []AddressInfo  `json:"rejectors"`
	GasTokenInfo          *TokenInfo     `json:"gasTokenInfo"`
	Trusted               bool           `json:"trusted"`
	Proposer              AddressInfo    `json:"proposer"`
	ProposedByDelegate    *string        `json:"proposedByDelegate"`
}

type Confirmation struct {
	Signer      AddressInfo `json:"signer"`
	Signature   string      `json:"signature"`
	SubmittedAt int64       `json:"submittedAt"`
}

type TokenInfo struct {
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
	LogoURI  string `json:"logoUri"`
}
