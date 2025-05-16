package safe

import (
	"context"
	"errors"
	"fmt"
	"math/big"
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
	log                   logrus.FieldLogger
	name                  string
	monitor               string
	ethereumPool          *ethereum.Pool
	address               string
	threshold             int
	signers               []string
	splitAddress          string
	splitsContractAddress string
	recoveryAccounts      []string
	recoveryAllocations   []uint32

	safeClient safe.Client

	excessQueue    *alert.ExcessQueue
	confirmations  *alert.Confirmations
	next           *alert.Next
	missing        *alert.Missing
	invalid        *alert.Invalid
	signersAlert   *alert.Signers
	thresholdAlert *alert.Threshold
	metrics        *Metrics

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
		threshold:             config.Threshold,
		signers:               config.Signers,
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
		signersAlert:          alert.NewSigners(log),
		thresholdAlert:        alert.NewThreshold(log, config.Threshold),
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
	go c.gatherMetrics(ctx)

	safeRsp, err := c.safeClient.GetSafe(ctx, c.address)
	if err != nil {
		c.log.WithError(err).Error("failed to get safe")

		return
	}

	// Check signers and threshold
	c.checkSignersAndThreshold(ctx, safeRsp)

	// Get and process queued transactions
	queuedTxData, err := c.getQueuedTransactions(ctx)
	if err != nil {
		return
	}

	// Process alerts based on transaction data
	c.processTransactionAlerts(ctx, queuedTxData)

	// Update metrics based on transaction data
	c.updateTransactionMetrics(queuedTxData)
}

// checkSignersAndThreshold verifies signers and threshold match expectations
func (c *Safe) checkSignersAndThreshold(ctx context.Context, safeRsp *safe.SafeResponse) {
	// Check signers
	signersMatch := safeRsp.CheckSigners(c.signers)

	shouldAlert := c.signersAlert.Update(!signersMatch)
	if shouldAlert {
		c.log.Warn("Alerting signer mismatch")

		if pErr := c.publisher.Publish(event.NewSignerMismatch(time.Now(), c.monitor, c.name, c.address)); pErr != nil {
			c.log.WithError(pErr).Error("Error publishing signer mismatch alert")
		}
	}

	c.metrics.UpdateSignersValid(boolToFloat64(signersMatch), []string{c.name, c.address, c.Type()})

	// Check threshold
	shouldAlert = c.thresholdAlert.Update(safeRsp.Threshold)
	if shouldAlert {
		c.log.Warn("Alerting threshold mismatch")

		if pErr := c.publisher.Publish(event.NewThreshold(time.Now(), c.monitor, c.name, c.address, safeRsp.Threshold, c.threshold)); pErr != nil {
			c.log.WithError(pErr).Error("Error publishing threshold mismatch alert")
		}
	}

	c.metrics.UpdateThresholdValid(boolToFloat64(safeRsp.Threshold == c.threshold), []string{c.name, c.address, c.Type()})
}

// TransactionData holds information about queued transactions
type TransactionData struct {
	Transactions          []*safe.QueuedTransactionResult
	RecoveryTxID          string
	InvalidRecoveryError  error
	HasNextRecoveryTx     bool
	CurrentConfirmations  int
	RequiredConfirmations int
}

// getQueuedTransactions fetches and processes queued transactions
func (c *Safe) getQueuedTransactions(ctx context.Context) (*TransactionData, error) {
	queued, err := c.safeClient.GetQueuedTransactions(ctx, c.address)
	if err != nil {
		c.log.WithError(err).Error("failed to get queued transactions")

		return nil, err
	}

	var txns []*safe.QueuedTransactionResult

	for _, tx := range queued.Results {
		if tx.Type == "TRANSACTION" {
			txns = append(txns, &tx)
		}
	}

	c.metrics.UpdateTransactionQueueSize(float64(len(txns)), []string{c.name, c.address, c.Type()})

	txData := &TransactionData{
		Transactions: txns,
	}

	// Process transactions to find valid recovery transactions
	c.processTransactions(ctx, txData)

	return txData, nil
}

// processTransactions examines each transaction to identify recovery transactions
func (c *Safe) processTransactions(ctx context.Context, txData *TransactionData) {
	for i, tx := range txData.Transactions {
		txDetails, err := c.safeClient.GetTransaction(ctx, c.address, tx.Transaction.ID)
		if err != nil {
			c.log.WithError(err).Error("failed to get recovery transaction details")

			return
		}

		if !c.isValidRecoveryTransaction(tx.Transaction.ID, txDetails) {
			continue
		}

		txData.RecoveryTxID = tx.Transaction.ID
		txData.InvalidRecoveryError = nil

		if err := c.checkRecoveryParameters(txDetails); err != nil {
			c.log.WithFields(logrus.Fields{
				"tx_id": tx.Transaction.ID,
			}).WithError(err).Warn("invalid recovery transaction queued")

			txData.InvalidRecoveryError = err

			continue
		}

		if i == 0 {
			txData.HasNextRecoveryTx = true
		}

		if txData.RequiredConfirmations == 0 {
			txData.CurrentConfirmations = len(txDetails.DetailedExecutionInfo.Confirmations)
			txData.RequiredConfirmations = txDetails.DetailedExecutionInfo.ConfirmationsRequired
		}
	}
}

// isValidRecoveryTransaction checks if a transaction is a valid recovery transaction
func (c *Safe) isValidRecoveryTransaction(txID string, txDetails *safe.TransactionDetails) bool {
	// Check if txDetails is nil
	if txDetails == nil {
		c.log.WithFields(logrus.Fields{
			"tx_id": txID,
		}).WithError(errors.New("transaction details is nil")).Warn("invalid recovery transaction queued")

		return false
	}

	// Check to address
	if txDetails.TxData.To.Value != c.splitsContractAddress {
		c.log.WithFields(logrus.Fields{
			"tx_id": txID,
		}).WithError(errors.New("invalid to address")).Warn("non-split recovery transaction queued")

		return false
	}

	// Check if data decoded is nil
	if txDetails.TxData.DataDecoded == nil {
		c.log.WithFields(logrus.Fields{
			"tx_id": txID,
		}).WithError(errors.New("data decoded is nil")).Warn("non-split recovery transaction queued")

		return false
	}

	// Check method
	if txDetails.TxData.DataDecoded.Method != "updateSplit" {
		c.log.WithFields(logrus.Fields{
			"tx_id":  txID,
			"method": txDetails.TxData.DataDecoded.Method,
		}).WithError(errors.New("invalid method name, should be updateSplit")).Warn("non-split recovery transaction queued")

		return false
	}

	return true
}

// processTransactionAlerts processes various alerts based on transaction data
func (c *Safe) processTransactionAlerts(ctx context.Context, txData *TransactionData) {
	// Check if queue size is too large
	c.alertQueueExcess(ctx, len(txData.Transactions))

	// Check for missing recovery transaction
	c.alertMissingRecovery(ctx, txData.RecoveryTxID)

	// Check for invalid recovery transaction
	c.alertInvalidRecovery(ctx, txData.RecoveryTxID, txData.InvalidRecoveryError)

	// Check if recovery transaction is not next in queue
	c.alertNotNextRecovery(ctx, txData)

	// Check confirmation status
	c.alertConfirmationStatus(ctx, txData)
}

// alertQueueExcess checks and alerts if queue size is too large
func (c *Safe) alertQueueExcess(ctx context.Context, queueSize int) {
	shouldAlert := c.excessQueue.Update(queueSize)
	if shouldAlert {
		c.log.WithFields(logrus.Fields{
			"length": queueSize,
		}).Warn("Alerting transaction queue size")

		if err := c.publisher.Publish(event.NewTransactionQueueExcess(time.Now(), c.monitor, c.name, c.address, queueSize)); err != nil {
			c.log.WithError(err).WithField("length", queueSize).Error("Error publishing transaction queue excess alert")
		}
	}
}

// alertMissingRecovery checks and alerts if recovery transaction is missing
func (c *Safe) alertMissingRecovery(ctx context.Context, recoveryTx string) {
	shouldAlert := c.missing.Update(recoveryTx == "")
	if shouldAlert {
		c.log.Warn("Alerting recovery transaction missing")

		if err := c.publisher.Publish(event.NewRecoveryTransactionMissing(time.Now(), c.monitor, c.name, c.address)); err != nil {
			c.log.WithError(err).WithField("tx_id", recoveryTx).Error("Error publishing recovery transaction missing alert")
		}
	}
}

// alertInvalidRecovery checks and alerts if recovery transaction is invalid
func (c *Safe) alertInvalidRecovery(ctx context.Context, recoveryTx string, invalidError error) {
	shouldAlert := c.invalid.Update(invalidError)
	if shouldAlert && invalidError != nil {
		c.log.WithFields(logrus.Fields{
			"tx_id": recoveryTx,
		}).WithError(invalidError).Warn("Alerting recovery transaction invalid")

		if err := c.publisher.Publish(event.NewRecoveryTransactionInvalid(time.Now(), c.monitor, c.name, c.address, recoveryTx, invalidError.Error())); err != nil {
			c.log.WithError(err).WithField("tx_id", recoveryTx).Error("Error publishing recovery transaction invalid alert")
		}
	}
}

// alertNotNextRecovery checks and alerts if recovery transaction is not next in queue
func (c *Safe) alertNotNextRecovery(ctx context.Context, txData *TransactionData) {
	shouldAlert := c.next.Update(txData.RecoveryTxID != "", txData.InvalidRecoveryError == nil, txData.HasNextRecoveryTx)
	if shouldAlert {
		c.log.WithFields(logrus.Fields{
			"tx_id": txData.RecoveryTxID,
		}).Warn("Alerting recovery transaction not next")

		if err := c.publisher.Publish(event.NewRecoveryTransactionNotNext(time.Now(), c.monitor, c.name, c.address, txData.RecoveryTxID)); err != nil {
			c.log.WithError(err).WithField("tx_id", txData.RecoveryTxID).Error("Error publishing recovery transaction not next alert")
		}
	}
}

// alertConfirmationStatus checks and alerts about confirmation status
func (c *Safe) alertConfirmationStatus(ctx context.Context, txData *TransactionData) {
	expectedConfirmations := txData.RequiredConfirmations - 1
	// Handle special case where a safe multisig only requires 1 confirmation
	if txData.RequiredConfirmations == 1 {
		expectedConfirmations = 1
	}

	shouldAlert := c.confirmations.Update(txData.CurrentConfirmations, expectedConfirmations, txData.HasNextRecoveryTx)
	if shouldAlert {
		c.log.WithFields(logrus.Fields{
			"current_confirmations":  txData.CurrentConfirmations,
			"expected_confirmations": expectedConfirmations,
			"tx_id":                  txData.RecoveryTxID,
		}).Warn("Alerting recovery transaction not pre-signed")

		if err := c.publisher.Publish(event.NewRecoveryTransactionConfirmations(time.Now(), c.monitor, c.name, c.address, txData.RecoveryTxID, txData.CurrentConfirmations, expectedConfirmations)); err != nil {
			c.log.WithError(err).WithFields(logrus.Fields{
				"current_confirmations":  txData.CurrentConfirmations,
				"expected_confirmations": expectedConfirmations,
			}).Error("Error publishing recovery transaction confirmations alert")
		}
	}
}

// updateTransactionMetrics updates metrics based on transaction data
func (c *Safe) updateTransactionMetrics(txData *TransactionData) {
	// Calculate expected confirmations
	expectedConfirmations := txData.RequiredConfirmations - 1
	if txData.RequiredConfirmations == 1 {
		expectedConfirmations = 1
	}

	// Update metrics
	c.metrics.UpdateTransactionRecoveryValid(
		boolToFloat64(txData.RecoveryTxID != "" && txData.InvalidRecoveryError == nil),
		[]string{c.name, c.address, c.Type()},
	)
	c.metrics.UpdateTransactionRecoveryExists(
		boolToFloat64(txData.RecoveryTxID != ""),
		[]string{c.name, c.address, c.Type()},
	)
	c.metrics.UpdateTransactionRecoveryNext(
		boolToFloat64(txData.HasNextRecoveryTx),
		[]string{c.name, c.address, c.Type()},
	)

	c.metrics.ClearTransactionRecoveryPreSigned([]string{c.name, c.address, c.Type()})
	c.metrics.UpdateTransactionRecoveryPreSigned(
		boolToFloat64(txData.CurrentConfirmations == expectedConfirmations),
		[]string{c.name, c.address, c.Type(), strconv.Itoa(expectedConfirmations), strconv.Itoa(txData.CurrentConfirmations)},
	)
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}

	return 0
}

// TransactionParams contains parsed transaction parameters
type TransactionParams struct {
	SplitAddress   string
	Accounts       []string
	Allocations    []uint32
	DistributorFee uint32
}

func (c *Safe) checkRecoveryParameters(tx *safe.TransactionDetails) error {
	// Validate transaction basics
	if err := c.validateTxBasics(tx); err != nil {
		return err
	}

	// Parse parameters
	params, err := c.parseTxParameters(tx.TxData.DataDecoded.Parameters)
	if err != nil {
		return err
	}

	// Validate split address
	if !strings.EqualFold(params.SplitAddress, c.splitAddress) {
		return fmt.Errorf("invalid split address: got %s, want %s", params.SplitAddress, c.splitAddress)
	}

	// Validate array lengths
	if err := c.validateArrayLengths(params.Accounts, params.Allocations); err != nil {
		return err
	}

	// Validate accounts and allocations
	if err := c.validateAccountsAndAllocations(params.Accounts, params.Allocations); err != nil {
		return err
	}

	// Validate distributor fee
	if params.DistributorFee != 0 {
		return fmt.Errorf("distributor fee must be 0")
	}

	return nil
}

// validateTxBasics performs basic validation on the transaction
func (c *Safe) validateTxBasics(tx *safe.TransactionDetails) error {
	// First, verify tx and its data are not nil
	if tx == nil {
		return fmt.Errorf("transaction details is nil")
	}

	if tx.TxData.DataDecoded == nil {
		return fmt.Errorf("transaction data decoded is nil")
	}

	if tx.TxData.DataDecoded.Parameters == nil {
		return fmt.Errorf("transaction parameters are nil")
	}

	return nil
}

// parseTxParameters parses transaction parameters
func (c *Safe) parseTxParameters(parameters []safe.Parameter) (*TransactionParams, error) {
	result := &TransactionParams{}

	for _, param := range parameters {
		if param.Value == nil {
			return nil, fmt.Errorf("parameter %s has nil value", param.Name)
		}

		switch param.Name {
		case "split":
			if err := c.parseSplitAddress(param, result); err != nil {
				return nil, err
			}
		case "accounts":
			if err := c.parseAccounts(param, result); err != nil {
				return nil, err
			}
		case "percentAllocations":
			if err := c.parseAllocations(param, result); err != nil {
				return nil, err
			}
		case "distributorFee":
			if err := c.parseDistributorFee(param, result); err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

// parseSplitAddress parses the split address parameter
func (c *Safe) parseSplitAddress(param safe.Parameter, result *TransactionParams) error {
	var ok bool

	result.SplitAddress, ok = param.Value.(string)
	if !ok {
		return fmt.Errorf("invalid split value: expected string, got %T (%v)", param.Value, param.Value)
	}

	return nil
}

// parseAccounts parses the accounts parameter
func (c *Safe) parseAccounts(param safe.Parameter, result *TransactionParams) error {
	accountsIface, ok := param.Value.([]interface{})
	if !ok {
		return fmt.Errorf("invalid accounts value: expected []interface{}, got %T (%v)", param.Value, param.Value)
	}

	result.Accounts = make([]string, len(accountsIface))

	for i, acc := range accountsIface {
		if acc == nil {
			return fmt.Errorf("account at index %d is nil", i)
		}

		result.Accounts[i], ok = acc.(string)
		if !ok {
			return fmt.Errorf("invalid account value at index %d: expected string, got %T (%v)", i, acc, acc)
		}
	}

	return nil
}

// parseAllocations parses the percentage allocations parameter
func (c *Safe) parseAllocations(param safe.Parameter, result *TransactionParams) error {
	allocsIface, ok := param.Value.([]interface{})
	if !ok {
		return fmt.Errorf("invalid percentAllocations value: expected []interface{}, got %T (%v)", param.Value, param.Value)
	}

	result.Allocations = make([]uint32, len(allocsIface))

	for i, a := range allocsIface {
		if a == nil {
			return fmt.Errorf("allocation at index %d is nil", i)
		}

		aStr, ok := a.(string)
		if !ok {
			return fmt.Errorf("invalid allocation value type at index %d: expected string, got %T (%v)", i, a, a)
		}

		val, err := strconv.ParseUint(aStr, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid allocation value at index %d: %v", i, err)
		}

		result.Allocations[i] = uint32(val)
	}

	return nil
}

// parseDistributorFee parses the distributor fee parameter
func (c *Safe) parseDistributorFee(param safe.Parameter, result *TransactionParams) error {
	feeStr, ok := param.Value.(string)
	if !ok {
		return fmt.Errorf("invalid distributor fee value type: expected string, got %T (%v)", param.Value, param.Value)
	}

	val, err := strconv.ParseUint(feeStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid distributor fee value: %v", err)
	}

	result.DistributorFee = uint32(val)

	return nil
}

// validateArrayLengths validates that all arrays have consistent lengths
func (c *Safe) validateArrayLengths(accounts []string, allocations []uint32) error {
	// Check for nil slices
	if c.recoveryAccounts == nil {
		return fmt.Errorf("recovery accounts is nil")
	}

	if c.recoveryAllocations == nil {
		return fmt.Errorf("recovery allocations is nil")
	}

	// Check array lengths match
	if len(accounts) != len(c.recoveryAccounts) {
		return fmt.Errorf("invalid number of accounts: got %d, want %d", len(accounts), len(c.recoveryAccounts))
	}

	// Check allocations array length matches accounts
	if len(allocations) != len(accounts) {
		return fmt.Errorf("mismatched lengths: accounts (%d) and allocations (%d)", len(accounts), len(allocations))
	}

	// Check recovery allocations length matches recovery accounts
	if len(c.recoveryAllocations) != len(c.recoveryAccounts) {
		return fmt.Errorf("mismatched lengths: recovery accounts (%d) and recovery allocations (%d)",
			len(c.recoveryAccounts), len(c.recoveryAllocations))
	}

	return nil
}

// validateAccountsAndAllocations validates that accounts and allocations match expected values
func (c *Safe) validateAccountsAndAllocations(accounts []string, allocations []uint32) error {
	// Now we can safely iterate, knowing all arrays have consistent lengths
	for i, acc := range c.recoveryAccounts {
		if i >= len(accounts) {
			// This should not happen due to earlier checks, but defensively check anyway
			return fmt.Errorf("account index %d out of bounds (max %d)", i, len(accounts)-1)
		}

		if !strings.EqualFold(acc, accounts[i]) {
			return fmt.Errorf("invalid account at position %d: got %s, want %s",
				i, accounts[i], acc)
		}

		if i >= len(allocations) || i >= len(c.recoveryAllocations) {
			// This should not happen due to earlier checks, but defensively check anyway
			return fmt.Errorf("allocation index %d out of bounds", i)
		}

		if allocations[i] != c.recoveryAllocations[i] {
			return fmt.Errorf("invalid allocation for %s: got %d, want %d",
				acc, allocations[i], c.recoveryAllocations[i])
		}
	}

	return nil
}

func (c *Safe) gatherMetrics(ctx context.Context) {
	for _, node := range c.ethereumPool.GetHealthyExecutionNodes() {
		balance, err := node.BalanceAt(ctx, c.address)
		if err != nil {
			c.log.WithError(err).WithField("node", node.Name()).Error("Error fetching balance")
		}

		if balance == nil {
			c.log.WithField("node", node.Name()).Error("Balance is nil")

			continue
		}

		balanceFloat := new(big.Float).SetInt(balance)
		balanceFloat64, _ := balanceFloat.Float64()

		c.metrics.UpdateBalance(balanceFloat64, []string{c.name, c.Type(), node.Name(), c.address})
	}
}
