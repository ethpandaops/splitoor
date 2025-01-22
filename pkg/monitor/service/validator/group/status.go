package group

import (
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/ethpandaops/splitoor/pkg/monitor/beaconchain"
)

type MetricsStatus string

const (
	MetricsStatusUnknown         MetricsStatus = "unknown"
	MetricsStatusMempool         MetricsStatus = "mempool"
	MetricsStatusDeposited       MetricsStatus = "deposited"
	MetricsStatusPending         MetricsStatus = "pending"
	MetricsStatusDepositInvalid  MetricsStatus = "deposit_invalid"
	MetricsStatusActiveOnline    MetricsStatus = "active_online"
	MetricsStatusActiveOffline   MetricsStatus = "active_offline"
	MetricsStatusExitingOnline   MetricsStatus = "exiting_online"
	MetricsStatusExitingOffline  MetricsStatus = "exiting_offline"
	MetricsStatusSlashingOnline  MetricsStatus = "slashing_online"
	MetricsStatusSlashingOffline MetricsStatus = "slashing_offline"
	MetricsStatusExitedUnslashed MetricsStatus = "exited_unslashed"
	MetricsStatusExitedSlashed   MetricsStatus = "exited_slashed"
)

// Current statuses no supported by BeaconAPI
//   - mempool
//   - deposit_invalid
//   - active_offline
//   - exiting_offline
//   - slashing_offline
//
// offline could be supported by checking the last 3 epochs
func BeaconAPIToMetricsStatus(status v1.ValidatorState, slashed bool) MetricsStatus {
	switch status {
	case v1.ValidatorStatePendingInitialized:
		return MetricsStatusDeposited
	case v1.ValidatorStatePendingQueued:
		return MetricsStatusPending
	case v1.ValidatorStateActiveOngoing:
		return MetricsStatusActiveOnline
	case v1.ValidatorStateActiveExiting:
		return MetricsStatusExitingOnline
	case v1.ValidatorStateActiveSlashed:
		return MetricsStatusSlashingOnline
	case v1.ValidatorStateExitedUnslashed:
		return MetricsStatusExitedUnslashed
	case v1.ValidatorStateExitedSlashed:
		return MetricsStatusExitedSlashed
	case v1.ValidatorStateWithdrawalPossible,
		v1.ValidatorStateWithdrawalDone:
		if slashed {
			return MetricsStatusExitedSlashed
		}

		return MetricsStatusExitedUnslashed
	default:
		return MetricsStatusUnknown
	}
}

func BeaconchainToMetricsStatus(status beaconchain.Status) MetricsStatus {
	switch status {
	case beaconchain.StatusMempool:
		return MetricsStatusMempool
	case beaconchain.StatusDeposited:
		return MetricsStatusDeposited
	case beaconchain.StatusPending:
		return MetricsStatusPending
	case beaconchain.StatusDepositInvalid:
		return MetricsStatusDepositInvalid
	case beaconchain.StatusActiveOnline:
		return MetricsStatusActiveOnline
	case beaconchain.StatusActiveOffline:
		return MetricsStatusActiveOffline
	case beaconchain.StatusExitingOnline:
		return MetricsStatusExitingOnline
	case beaconchain.StatusExitingOffline:
		return MetricsStatusExitingOffline
	case beaconchain.StatusSlashingOnline:
		return MetricsStatusSlashingOnline
	case beaconchain.StatusSlashingOffline:
		return MetricsStatusSlashingOffline
	case beaconchain.StatusSlashed:
		return MetricsStatusExitedSlashed
	case beaconchain.StatusExited:
		return MetricsStatusExitedUnslashed
	default:
		return MetricsStatusUnknown
	}
}
