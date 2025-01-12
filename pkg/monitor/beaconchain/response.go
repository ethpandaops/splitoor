package beaconchain

type Response[T any] struct {
	Data   T      `json:"data"`
	Status string `json:"status"`
}

// enum status
type Status string

const (
	// Transaction is in mempool waiting to be included in a block
	StatusMempool Status = "mempool"

	// Deposit transaction has been included in a block and is waiting for finalization
	StatusDeposited Status = "deposited"

	// Deposit is finalized and validator is in activation queue
	StatusPending Status = "pending"

	// The transaction had an invalid BLS signature
	StatusDepositInvalid Status = "deposit_invalid"

	// Validator is actively participating in consensus
	StatusActiveOnline Status = "active_online"

	// An active validator has not been attesting for at least two epochs
	StatusActiveOffline Status = "active_offline"

	// The validator is online and currently exiting the network because either its balance dropped below 16ETH (forced exit)
	// or the exit was requested (voluntary exit) by the validator
	StatusExitingOnline Status = "exiting_online"

	// The validator is offline and currently exiting the network because either its balance dropped below 16ETH
	// or the exit was requested by the validator
	StatusExitingOffline Status = "exiting_offline"

	// The validator is online but was malicious and therefore forced to exit the network
	StatusSlashingOnline Status = "slashing_online"

	// The validator is offline and was malicious and which lead to a forced to exit out of the network.
	// The validator is currently in the exiting queue with a minimum of 25 minutes
	StatusSlashingOffline Status = "slashing_offline"

	// The validator has been kicked out of the network. The funds will be withdrawable after 36 days
	StatusSlashed Status = "slashed"

	// The validator has exited the network. The funds will be withdrawable after 1 day
	StatusExited Status = "exited"
)

type Validator struct {
	ActivationEligibilityEpoch int    `json:"activationeligibilityepoch"`
	ActivationEpoch            int    `json:"activationepoch"`
	Balance                    int    `json:"balance"`
	EffectiveBalance           int    `json:"effectivebalance"`
	ExitEpoch                  int    `json:"exitepoch"`
	LastAttestationSlot        int    `json:"lastattestationslot"`
	Name                       string `json:"name"`
	Pubkey                     string `json:"pubkey"`
	Slashed                    bool   `json:"slashed"`
	Status                     Status `json:"status"`
	ValidatorIndex             int    `json:"validatorindex"`
	WithdrawableEpoch          int    `json:"withdrawableepoch"`
	WithdrawalCredentials      string `json:"withdrawalcredentials"`
	TotalWithdrawals           int    `json:"total_withdrawals"` //nolint:tagliatelle // beaconcha.in response
}

func (v *Validator) IsExited() bool {
	return v.Status == StatusSlashed || v.Status == StatusExited
}
