package group

import (
	"context"
	"encoding/hex"
	"sync"
	"time"

	"github.com/ethpandaops/splitoor/pkg/0xsplits/contract"
	spl "github.com/ethpandaops/splitoor/pkg/0xsplits/split"
	"github.com/ethpandaops/splitoor/pkg/ethereum"
	"github.com/ethpandaops/splitoor/pkg/monitor/notifier"
	"github.com/ethpandaops/splitoor/pkg/monitor/safe"
	"github.com/ethpandaops/splitoor/pkg/monitor/service/split/group/account"
	"github.com/ethpandaops/splitoor/pkg/monitor/service/split/group/alert"
	"github.com/ethpandaops/splitoor/pkg/monitor/service/split/group/controller"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	MinBalance                 uint64 = 31.95 * 1e9
	ExpectedStatuses                  = []string{"active_online", "active_offline"}
	WithdrawalCredentialsCodes        = []int64{1}
)

type Group struct {
	log     logrus.FieldLogger
	name    string
	monitor string

	publisher    *notifier.Publisher
	ethereumPool *ethereum.Pool
	safeClient   safe.Client

	address    string
	client     *spl.Client
	contract   string
	hash       string
	accounts   []*account.Account
	controller controller.Controller

	metrics *Metrics

	validatorState *State
	balanceAlerts  map[string]*alert.Controller
	statusAlerts   map[string]*alert.Hash

	mu sync.Mutex
}

func NewGroup(ctx context.Context, log logrus.FieldLogger, monitor string, conf *Config, ethereumPool *ethereum.Pool, publisher *notifier.Publisher, safeClient safe.Client) (*Group, error) {
	log = log.WithField("group", conf.Name)

	var c string

	if conf.Contract != nil {
		c = *conf.Contract
	}

	accounts := make([]*account.Account, len(conf.Accounts))

	for i, acc := range conf.Accounts {
		accounts[i] = account.NewAccount(log, monitor, conf.Name, acc.Address, acc.Allocation, acc.Monitor)
	}

	ctr, err := controller.NewController(ctx, log, monitor, conf.Name, conf.Controller.ControllerType, conf.Controller.Config, conf.Address, c, ethereumPool, safeClient, publisher)
	if err != nil {
		return nil, err
	}

	return &Group{
		log:            log,
		name:           conf.Name,
		monitor:        monitor,
		publisher:      publisher,
		ethereumPool:   ethereumPool,
		safeClient:     safeClient,
		address:        conf.Address,
		contract:       c,
		accounts:       accounts,
		controller:     ctr,
		metrics:        GetMetricsInstance("splitoor_split", monitor),
		balanceAlerts:  make(map[string]*alert.Controller),
		statusAlerts:   make(map[string]*alert.Hash),
		validatorState: NewState(log),
	}, nil
}

func (g *Group) Start(ctx context.Context) error {
	if g.controller != nil {
		if err := g.controller.Start(ctx); err != nil {
			return err
		}
	}

	for _, account := range g.accounts {
		if err := account.Start(ctx); err != nil {
			return err
		}
	}

	if err := g.setupSplit(ctx); err != nil {
		return err
	}

	g.tick(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second * 12):
				g.tick(ctx)
			}
		}
	}()

	return nil
}

func (g *Group) Stop(ctx context.Context) error {
	if g.controller != nil {
		if err := g.controller.Stop(ctx); err != nil {
			return err
		}
	}

	for _, account := range g.accounts {
		if err := account.Stop(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (g *Group) setupSplit(ctx context.Context) error {
	log := g.log.WithField("split", g.name)

	if g.contract == "" {
		log.Debug("no contract address provided for split, requesting default contract address")

		dpNode, err := g.ethereumPool.WaitForHealthyExecutionNode(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to get healthy execution node to figure out default contract address")
		}

		chainID, err := dpNode.ChainID(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to get chain id from execution node")
		}

		address := contract.GetDefaultContractAddress(chainID.String())
		if address == nil {
			return errors.New("failed to get default contract address for chain id " + chainID.String())
		}

		g.contract = *address
	}

	sCfg := &spl.Config{
		ContractAddress: g.contract,
		SplitAddress:    &g.address,
	}

	var err error

	g.client, err = spl.NewClient(log, sCfg)
	if err != nil {
		return errors.Wrap(err, "failed to create split client")
	}

	accounts := []string{}
	allocations := []uint32{}

	for _, account := range g.accounts {
		accounts = append(accounts, account.Address())
		allocations = append(allocations, uint32(account.Allocation()))
	}

	hashParams := &spl.HashParams{
		Accounts:              accounts,
		PercentageAllocations: allocations,
	}

	hash, err := g.client.CalculateHash(hashParams)
	if err != nil {
		return errors.Wrap(err, "failed to calculate hash")
	}

	g.hash = hex.EncodeToString(hash)

	return nil
}

func (g *Group) tick(ctx context.Context) {
	// collect account balance
	// collect split balance eth
	// collect split balance of accounts with monitor (getETHBalance)
	// check split contract controller
	// check split contract hash

	// newState := NewState(g.log)

	// var wg sync.WaitGroup

	// if g.beaconchain != nil && time.Since(g.beaconchainLastTick) > g.beaconchain.GetCheckInterval() {
	// 	wg.Add(1)

	// 	go func() {
	// 		defer wg.Done()
	// 		g.tickBeaconchain(ctx, newState)
	// 	}()
	// }

	// if g.ethereumPool != nil && g.ethereumPool.HasHealthyBeaconNodes() {
	// 	wg.Add(1)

	// 	go func() {
	// 		defer wg.Done()
	// 		g.tickBeaconAPI(ctx, newState)
	// 	}()
	// }

	// wg.Wait()

	// g.mu.Lock()
	// changedPubkeys := g.validatorState.Merge(newState)
	// g.mu.Unlock()

	// g.updateAlerts(changedPubkeys)
}

// func (g *Group) tickBeaconAPI(ctx context.Context, state *State) {
// 	for _, node := range g.ethereumPool.GetHealthyBeaconNodes() {
// 		validators, err := node.Node().FetchValidators(ctx, "head", nil, g.pubkeys)
// 		if err != nil {
// 			g.log.WithError(err).WithField("node", node.Name()).Error("Error fetching validators")
// 		}

// 		for _, validator := range validators {
// 			g.updateValidatorBeaconAPI(validator, node.Name(), state)
// 		}
// 	}
// }

// func (g *Group) tickBeaconchain(ctx context.Context, state *State) {
// 	g.log.Debug("Starting beaconchain validators update")

// 	g.beaconchainLastTick = time.Now()

// 	chunkIndex := 0

// 	for chunkIndex < len(g.beaconchainChunks) {
// 		// Process up to max requests per minute
// 		requestCount := 0
// 		for requestCount < g.beaconchain.GetMaxRequestsPerMinute() && chunkIndex < len(g.beaconchainChunks) {
// 			chunk := g.beaconchainChunks[chunkIndex]

// 			g.log.WithFields(logrus.Fields{
// 				"chunk":  chunkIndex,
// 				"length": len(chunk),
// 			}).Debug("Processing beaconchain validator pubkeys chunk")

// 			err := g.getValidatorsBeaconchain(ctx, chunk, state)
// 			if err != nil {
// 				g.log.WithError(err).WithField("pubkeys", chunk).Error("Error updating beaconchain validators")
// 			}

// 			requestCount++
// 			chunkIndex++
// 		}

// 		// Wait a minute after the last request before continuing
// 		if requestCount > 0 && chunkIndex < len(g.beaconchainChunks) {
// 			timer := time.NewTimer(time.Minute)
// 			select {
// 			case <-timer.C:
// 			case <-ctx.Done():
// 				timer.Stop()

// 				return
// 			}
// 		}
// 	}

// 	g.log.Debug("Finished beaconchain validators update")
// }

// func (g *Group) getValidatorsBeaconchain(ctx context.Context, validators []string, state *State) error {
// 	if len(validators) == 0 {
// 		return nil
// 	}

// 	if len(validators) == 1 {
// 		response, err := g.beaconchain.GetValidator(ctx, validators[0])
// 		if err != nil {
// 			return err
// 		}

// 		g.updateValidatorBeaconchain(response, state)
// 	} else {
// 		response, err := g.beaconchain.GetValidators(ctx, validators)
// 		if err != nil {
// 			return err
// 		}

// 		for _, validator := range response {
// 			if validator != nil {
// 				g.updateValidatorBeaconchain(validator, state)
// 			}
// 		}
// 	}

// 	return nil
// }

// func (g *Group) updateValidatorBeaconchain(data *beaconchain.Validator, state *State) {
// 	if data == nil {
// 		return
// 	}

// 	source := "beaconcha.in"
// 	labels := []string{
// 		g.name,
// 		data.Pubkey,
// 		source,
// 	}

// 	g.metrics.UpdateBalance(float64(data.Balance), labels)

// 	status := BeaconchainToMetricsStatus(data.Status)

// 	credentialsCode, err := GetWithdrawalCredentialsCode(data.WithdrawalCredentials)
// 	if err != nil {
// 		g.log.WithError(err).WithField("credentials", data.WithdrawalCredentials).Error("Error parsing withdrawal credentials")
// 	}

// 	code := float64(0)
// 	if credentialsCode != nil {
// 		code = float64(*credentialsCode)

// 		state.UpdateValidator(source, data.Pubkey, uint64(data.Balance), status, *credentialsCode)
// 	}

// 	g.metrics.UpdateCredentialsCode(code, labels)
// 	g.metrics.UpdateLastAttestationSlot(float64(data.LastAttestationSlot), labels)
// 	g.metrics.UpdateTotalWithdrawals(float64(data.TotalWithdrawals), labels)
// 	g.metrics.UpdateStatus(status, labels)
// }

// func (g *Group) updateValidatorBeaconAPI(data *v1.Validator, source string, state *State) {
// 	if data == nil {
// 		return
// 	}

// 	val := data.Validator

// 	if val == nil {
// 		return
// 	}

// 	labels := []string{
// 		g.name,
// 		val.PublicKey.String(),
// 		source,
// 	}

// 	g.metrics.UpdateBalance(float64(data.Balance), labels)

// 	status := BeaconAPIToMetricsStatus(data.Status, val.Slashed)

// 	credentialsCode, err := GetWithdrawalCredentialsCode(hex.EncodeToString(val.WithdrawalCredentials))
// 	if err != nil {
// 		g.log.WithError(err).WithField("credentials", val.WithdrawalCredentials).Error("Error parsing withdrawal credentials")
// 	}

// 	code := float64(0)

// 	if credentialsCode != nil {
// 		state.UpdateValidator(source, val.PublicKey.String(), uint64(data.Balance), status, *credentialsCode)

// 		code = float64(*credentialsCode)
// 	}

// 	g.metrics.UpdateCredentialsCode(code, labels)
// 	g.metrics.UpdateCredentialsCode(code, labels)
// 	g.metrics.UpdateStatus(status, labels)
// }

// func (g *Group) updateAlerts(changedPubkeys []string) {
// 	g.mu.Lock()
// 	defer g.mu.Unlock()

// 	for _, pubkey := range changedPubkeys {
// 		if _, exists := g.balanceAlerts[pubkey]; !exists {
// 			g.balanceAlerts[pubkey] = alert.NewBalance(g.log, MinBalance)
// 		}

// 		if _, exists := g.statusAlerts[pubkey]; !exists {
// 			g.statusAlerts[pubkey] = alert.NewStatus(g.log, ExpectedStatuses)
// 		}

// 		if _, exists := g.withdrawalCredentialsAlerts[pubkey]; !exists {
// 			g.withdrawalCredentialsAlerts[pubkey] = alert.NewWithdrawalCredentials(g.log, WithdrawalCredentialsCodes)
// 		}

// 		balanceAlert := g.balanceAlerts[pubkey]
// 		statusAlert := g.statusAlerts[pubkey]
// 		withdrawalCredentialsAlert := g.withdrawalCredentialsAlerts[pubkey]

// 		balances := make([]uint64, 0, len(g.validatorState.Validators[pubkey].Sources))
// 		statuses := make([]string, 0, len(g.validatorState.Validators[pubkey].Sources))
// 		codes := make([]int64, 0, len(g.validatorState.Validators[pubkey].Sources))

// 		for _, source := range g.validatorState.Validators[pubkey].Sources {
// 			balances = append(balances, source.Balance)
// 			statuses = append(statuses, string(source.Status))
// 			codes = append(codes, source.WithdrawalCredentialsCode)
// 		}

// 		if shouldAlert, balance := balanceAlert.Update(balances); shouldAlert {
// 			g.log.WithField("balance", *balance).WithField("pubkey", pubkey).Warn("Alerting min balance")

// 			if err := g.publisher.Publish(validator.NewMinBalance(time.Now(), *balance, pubkey, g.name, g.monitor)); err != nil {
// 				g.log.WithError(err).WithField("pubkey", pubkey).Error("Error publishing min balance alert")
// 			}
// 		}

// 		if shouldAlert, alertingStatus := statusAlert.Update(statuses); shouldAlert {
// 			g.log.WithField("status", *alertingStatus).WithField("pubkey", pubkey).Warn("Alerting status")

// 			if err := g.publisher.Publish(validator.NewStatus(time.Now(), *alertingStatus, pubkey, g.name, g.monitor)); err != nil {
// 				g.log.WithError(err).WithField("pubkey", pubkey).WithField("status", *alertingStatus).Error("Error publishing status alert")
// 			}
// 		}

// 		if shouldAlert, alertingCredential := withdrawalCredentialsAlert.Update(codes); shouldAlert {
// 			g.log.WithField("credential", *alertingCredential).WithField("pubkey", pubkey).Warn("Alerting withdrawal credentials")

// 			if err := g.publisher.Publish(validator.NewWithdrawalCredentials(time.Now(), *alertingCredential, pubkey, g.name, g.monitor)); err != nil {
// 				g.log.WithError(err).WithField("pubkey", pubkey).WithField("credential", *alertingCredential).Error("Error publishing withdrawal credentials alert")
// 			}
// 		}
// 	}
// }

// func GetWithdrawalCredentialsCode(withdrawalCredentials string) (*int64, error) {
// 	if strings.HasPrefix(withdrawalCredentials, "0x") {
// 		i64, err := strconv.ParseInt(withdrawalCredentials[:4], 0, 64)
// 		if err != nil {
// 			return nil, err
// 		}

// 		return &i64, nil
// 	}

// 	i64, err := strconv.ParseInt(withdrawalCredentials[:2], 0, 64)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &i64, nil
// }
