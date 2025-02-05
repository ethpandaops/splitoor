package safe

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ethpandaops/splitoor/pkg/0xsplits/split"
	"github.com/ethpandaops/splitoor/pkg/ethereum"
	event "github.com/ethpandaops/splitoor/pkg/monitor/event/safe"
	"github.com/ethpandaops/splitoor/pkg/monitor/notifier"
	"github.com/ethpandaops/splitoor/pkg/monitor/safe"
	"github.com/ethpandaops/splitoor/pkg/monitor/service/split/group/controller/safe/alert"

	"github.com/sirupsen/logrus"
)

const (
	ControllerType = "safe"
	MaxQueueSize   = 1
)

type Safe struct {
	log           logrus.FieldLogger
	name          string
	monitor       string
	ethereumPool  *ethereum.Pool
	address       string
	minSignatures int

	splitAddress          string
	splitsContractAddress string
	recoveryAccounts      []string
	recoveryAllocations   []uint32

	safeClient safe.Client

	excessQueue   *alert.ExcessQueue
	confirmations *alert.Confirmations
	next          *alert.Next
	missing       *alert.Missing
	invalid       *alert.Invalid

	metrics *Metrics

	publisher *notifier.Publisher
}

func New(ctx context.Context, log logrus.FieldLogger, monitor, name string, config *Config, splitAddress, recoveryAddress, splitsContractAddress string, ethereumPool *ethereum.Pool, safeClient safe.Client, publisher *notifier.Publisher) (*Safe, error) {
	// expected recipients when split is in recovery state
	recoveryAccounts, recoveryAllocations, err := split.ParseRecipients([]string{splitAddress, recoveryAddress}, []uint32{1, 999999})
	if err != nil {
		return nil, err
	}

	return &Safe{
		log:                   log.WithField("controller", ControllerType).WithField("address", config.Address),
		name:                  name,
		monitor:               monitor,
		ethereumPool:          ethereumPool,
		address:               config.Address,
		minSignatures:         config.MinSignatures,
		splitAddress:          splitAddress,
		splitsContractAddress: splitsContractAddress,
		recoveryAccounts:      recoveryAccounts,
		recoveryAllocations:   recoveryAllocations,
		safeClient:            safeClient,
		excessQueue:           alert.NewExcessQueue(log, MaxQueueSize),
		confirmations:         alert.NewConfirmations(log),
		next:                  alert.NewNext(log),
		missing:               alert.NewMissing(log),
		invalid:               alert.NewInvalid(log),
		metrics:               GetMetricsInstance("splitoor_split_controller", monitor),
		publisher:             publisher,
	}, nil
}

func (c *Safe) Start(ctx context.Context) error {
	if c.safeClient == nil {
		c.log.Warn("Safe config disabled, skipping")

		return nil
	}

	c.tick(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second * 12):
				c.tick(ctx)
			}
		}
	}()

	return nil
}

func (c *Safe) Stop(ctx context.Context) error {
	return nil
}

func (c *Safe) Type() string {
	return ControllerType
}

func (c *Safe) Name() string {
	return c.name
}

func (c *Safe) Address() string {
	return c.address
}

func (c *Safe) tick(ctx context.Context) {
	queued, err := c.safeClient.GetQueuedTransactions(ctx, c.address)
	if err != nil {
		c.log.WithError(err).Error("failed to get queued transactions")

		return
	}

	var txns []*safe.QueuedTransactionResult

	for _, tx := range queued.Results {
		if tx.Type == "TRANSACTION" {
			txns = append(txns, &tx)
		}
	}

	c.metrics.UpdateTransactionQueueSize(float64(len(txns)), []string{c.name, c.address, c.Type()})

	// a tx with calling "updateSplit" on expected contract exists
	var recoveryTx string

	var invalidRecoveryError error

	// has a valid recovery tx next in the queue
	hasNextRecoveryTx := false

	// current number of confirmations of the first recovery tx
	var currentConfirmations int

	// required number of confirmations of the first recovery tx
	var requiredConfirmations int

	for i, tx := range txns {
		txDetails, err := c.safeClient.GetTransaction(ctx, tx.Transaction.ID)
		if err != nil {
			c.log.WithError(err).Error("failed to get recovery transaction details")

			return
		}

		// check to
		if txDetails.TxData.To.Value != c.splitsContractAddress {
			c.log.WithFields(logrus.Fields{
				"tx_id": tx.Transaction.ID,
			}).WithError(errors.New("invalid to address")).Warn("non-split recovery transaction queued")

			continue
		}

		if txDetails.TxData.DataDecoded == nil || txDetails.TxData.DataDecoded.Method != "updateSplit" {
			c.log.WithFields(logrus.Fields{
				"tx_id": tx.Transaction.ID,
			}).WithError(errors.New("invalid method name, should be updateSplit")).Warn("non-split recovery transaction queued")

			continue
		}

		recoveryTx = tx.Transaction.ID

		// clear previous error if another recovery tx exists
		invalidRecoveryError = nil

		if err := c.checkRecoveryParameters(txDetails); err != nil {
			c.log.WithFields(logrus.Fields{
				"tx_id": tx.Transaction.ID,
			}).WithError(err).Warn("invalid recovery transaction queued")

			invalidRecoveryError = err

			continue
		}

		if i == 0 {
			hasNextRecoveryTx = true
		}

		if requiredConfirmations == 0 {
			currentConfirmations = len(txDetails.DetailedExecutionInfo.Confirmations)
			requiredConfirmations = txDetails.DetailedExecutionInfo.ConfirmationsRequired
		}
	}

	/*
	 * Always alert if the queue is too large
	 */
	shouldAlert := c.excessQueue.Update(len(txns))
	if shouldAlert {
		c.log.WithFields(logrus.Fields{
			"length": len(txns),
		}).Warn("Alerting transaction queue size")

		if err := c.publisher.Publish(event.NewTransactionQueueExcess(time.Now(), c.monitor, c.name, c.address, len(txns))); err != nil {
			c.log.WithError(err).WithField("length", len(txns)).Error("Error publishing transaction queue excess alert")
		}
	}

	/*
	 * Alert if no valid or invalid recovery transaction exists
	 */
	shouldAlert = c.missing.Update(recoveryTx == "")
	if shouldAlert {
		c.log.Warn("Alerting recovery transaction missing")

		if err := c.publisher.Publish(event.NewRecoveryTransactionMissing(time.Now(), c.monitor, c.name, c.address)); err != nil {
			c.log.WithError(err).WithField("tx_id", recoveryTx).Error("Error publishing recovery transaction missing alert")
		}
	}

	/*
	 * Alert if an ivalid recovery transaction exists
	 */
	shouldAlert = c.invalid.Update(invalidRecoveryError)
	if shouldAlert {
		c.log.WithFields(logrus.Fields{
			"tx_id": recoveryTx,
		}).WithError(invalidRecoveryError).Warn("Alerting recovery transaction invalid")

		if err := c.publisher.Publish(event.NewRecoveryTransactionInvalid(time.Now(), c.monitor, c.name, c.address, recoveryTx, invalidRecoveryError.Error())); err != nil {
			c.log.WithError(err).WithField("tx_id", recoveryTx).Error("Error publishing recovery transaction invalid alert")
		}
	}

	/*
	 * Alert if a valid recovery transaction is not next in the queue
	 */
	shouldAlert = c.next.Update(recoveryTx != "", invalidRecoveryError == nil, hasNextRecoveryTx)
	if shouldAlert {
		c.log.WithFields(logrus.Fields{
			"tx_id": recoveryTx,
		}).Warn("Alerting recovery transaction not next")

		if err := c.publisher.Publish(event.NewRecoveryTransactionNotNext(time.Now(), c.monitor, c.name, c.address, recoveryTx)); err != nil {
			c.log.WithError(err).WithField("tx_id", recoveryTx).Error("Error publishing recovery transaction not next alert")
		}
	}

	expectedConfirmations := requiredConfirmations - 1
	// handle special case where a safe multisig only requires 1 confirmation
	if requiredConfirmations == 1 {
		expectedConfirmations = 1
	}

	/*
	 * Alert if a valid next recovery transaction is not pre-signed
	 */
	shouldAlert = c.confirmations.Update(currentConfirmations, expectedConfirmations, hasNextRecoveryTx)
	if shouldAlert {
		c.log.WithFields(logrus.Fields{
			"current_confirmations":  currentConfirmations,
			"expected_confirmations": expectedConfirmations,
			"tx_id":                  recoveryTx,
		}).Warn("Alerting recovery transaction not pre-signed")

		if err := c.publisher.Publish(event.NewRecoveryTransactionConfirmations(time.Now(), c.monitor, c.name, c.address, recoveryTx, currentConfirmations, expectedConfirmations)); err != nil {
			c.log.WithError(err).WithFields(logrus.Fields{
				"current_confirmations":  currentConfirmations,
				"expected_confirmations": expectedConfirmations,
			}).Error("Error publishing recovery transaction confirmations alert")
		}
	}

	c.metrics.UpdateTransactionRecoveryValid(boolToFloat64(recoveryTx != "" && invalidRecoveryError == nil), []string{c.name, c.address, c.Type()})
	c.metrics.UpdateTransactionRecoveryExists(boolToFloat64(recoveryTx != ""), []string{c.name, c.address, c.Type()})
	c.metrics.UpdateTransactionRecoveryNext(boolToFloat64(hasNextRecoveryTx), []string{c.name, c.address, c.Type()})
	c.metrics.UpdateTransactionRecoveryPreSigned(boolToFloat64(currentConfirmations == expectedConfirmations), []string{c.name, c.address, c.Type(), strconv.Itoa(expectedConfirmations), strconv.Itoa(currentConfirmations)})
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}

	return 0
}

func (c *Safe) checkRecoveryParameters(tx *safe.TransactionDetails) error {
	var splitAddress string

	var accounts []string

	var allocations []uint32

	var distributorFee uint32

	for _, param := range tx.TxData.DataDecoded.Parameters {
		switch param.Name {
		case "split":
			var ok bool

			splitAddress, ok = param.Value.(string)
			if !ok {
				return fmt.Errorf("invalid split value: %v", param.Value)
			}
		case "accounts":
			accountsIface, ok := param.Value.([]interface{})
			if !ok {
				return fmt.Errorf("invalid accounts value: %v", param.Value)
			}

			accounts = make([]string, len(accountsIface))

			for i, acc := range accountsIface {
				accounts[i], ok = acc.(string)
				if !ok {
					return fmt.Errorf("invalid account value: %v", acc)
				}
			}
		case "percentAllocations":
			allocsIface, ok := param.Value.([]interface{})
			if !ok {
				return fmt.Errorf("invalid percentAllocations value: %v", param.Value)
			}

			allocations = make([]uint32, len(allocsIface))

			for i, a := range allocsIface {
				val, err := strconv.ParseUint(a.(string), 10, 32)
				if err != nil {
					return fmt.Errorf("invalid allocation value: %v", err)
				}

				allocations[i] = uint32(val)
			}
		case "distributorFee":
			val, err := strconv.ParseUint(param.Value.(string), 10, 32)
			if err != nil {
				return fmt.Errorf("invalid distributor fee value: %v", err)
			}

			distributorFee = uint32(val)
		}
	}

	if !strings.EqualFold(splitAddress, c.splitAddress) {
		return fmt.Errorf("invalid split address: got %s, want %s", splitAddress, c.splitAddress)
	}

	if len(accounts) != len(c.recoveryAccounts) {
		return fmt.Errorf("invalid number of accounts: got %d, want %d", len(accounts), len(c.recoveryAccounts))
	}

	for i, acc := range c.recoveryAccounts {
		if !strings.EqualFold(acc, accounts[i]) {
			return fmt.Errorf("invalid account at position %d: got %s, want %s",
				i, accounts[i], acc)
		}

		if allocations[i] != c.recoveryAllocations[i] {
			return fmt.Errorf("invalid allocation for %s: got %d, want %d",
				acc, allocations[i], c.recoveryAllocations[i])
		}
	}

	if distributorFee != 0 {
		return fmt.Errorf("distributor fee must be 0")
	}

	return nil
}
