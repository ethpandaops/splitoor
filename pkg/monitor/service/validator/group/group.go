package group

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/ethpandaops/splitoor/pkg/ethereum"
	"github.com/ethpandaops/splitoor/pkg/monitor/beaconchain"
	"github.com/ethpandaops/splitoor/pkg/monitor/event/validator"
	"github.com/ethpandaops/splitoor/pkg/monitor/notifier"
	"github.com/ethpandaops/splitoor/pkg/monitor/service/validator/group/alert"
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

	publisher *notifier.Publisher

	ethereumPool *ethereum.Pool
	pubkeys      []phase0.BLSPubKey

	beaconchain         beaconchain.Client
	beaconchainChunks   [][]string
	beaconchainLastTick time.Time

	metrics *Metrics

	validatorState              *State
	balanceAlerts               map[string]*alert.Balance
	statusAlerts                map[string]*alert.Status
	withdrawalCredentialsAlerts map[string]*alert.WithdrawalCredentials
	mu                          sync.Mutex
}

func NewGroup(ctx context.Context, log logrus.FieldLogger, monitor string, conf *Config, ethereumPool *ethereum.Pool, bc beaconchain.Client, publisher *notifier.Publisher) (*Group, error) {
	var chunks [][]string

	if bc != nil {
		for i := 0; i < len(conf.Pubkeys); i += bc.GetBatchSize() {
			end := i + bc.GetBatchSize()
			if end > len(conf.Pubkeys) {
				end = len(conf.Pubkeys)
			}

			chunks = append(chunks, conf.Pubkeys[i:end])
		}
	}

	pubkeys := make([]phase0.BLSPubKey, len(conf.Pubkeys))

	for i, pubkey := range conf.Pubkeys {
		bytes, err := hex.DecodeString(strings.TrimPrefix(pubkey, "0x"))
		if err != nil {
			return nil, fmt.Errorf("invalid pubkey %s: %w", pubkey, err)
		}

		copy(pubkeys[i][:], bytes)
	}

	log = log.WithField("group", conf.Name)

	return &Group{
		log:                         log,
		name:                        conf.Name,
		monitor:                     monitor,
		publisher:                   publisher,
		ethereumPool:                ethereumPool,
		pubkeys:                     pubkeys,
		beaconchain:                 bc,
		beaconchainChunks:           chunks,
		metrics:                     GetMetricsInstance("splitoor_validator", monitor),
		balanceAlerts:               make(map[string]*alert.Balance),
		statusAlerts:                make(map[string]*alert.Status),
		withdrawalCredentialsAlerts: make(map[string]*alert.WithdrawalCredentials),
		validatorState:              NewState(log),
	}, nil
}

func (g *Group) Start(ctx context.Context) {
	g.tick(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 12):
			g.tick(ctx)
		}
	}
}

func (g *Group) Stop(ctx context.Context) error {
	return nil
}

func (g *Group) tick(ctx context.Context) {
	newState := NewState(g.log)

	var wg sync.WaitGroup

	if g.beaconchain != nil && time.Since(g.beaconchainLastTick) > g.beaconchain.GetCheckInterval() {
		wg.Add(1)

		go func() {
			defer wg.Done()
			g.checkBeaconchain(ctx, newState)
		}()
	}

	if g.ethereumPool != nil && g.ethereumPool.HasHealthyBeaconNodes() {
		wg.Add(1)

		go func() {
			defer wg.Done()
			g.checkBeaconAPI(ctx, newState)
		}()
	}

	wg.Wait()

	g.mu.Lock()
	changedPubkeys := g.validatorState.Merge(newState)
	g.mu.Unlock()

	g.updateAlerts(changedPubkeys)
}

func (g *Group) checkBeaconAPI(ctx context.Context, state *State) {
	for _, node := range g.ethereumPool.GetHealthyBeaconNodes() {
		validators, err := node.Node().FetchValidators(ctx, "head", nil, g.pubkeys)
		if err != nil {
			g.log.WithError(err).WithField("source", node.Name()).Error("Error fetching validators")

			for _, pubkey := range g.pubkeys {
				g.metrics.UpdateStatus(MetricsStatusUnknown, []string{g.name, pubkey.String(), node.Name()})
			}

			continue
		}

		foundPubkeys := make(map[string]bool)

		for _, validator := range validators {
			g.updateValidatorBeaconAPI(validator, node.Name(), state)

			pubkey, err := validator.PubKey(ctx)
			if err != nil {
				g.log.WithError(err).WithField("source", node.Name()).WithField("validator_index", validator.Index).Error("Error getting pubkey")

				continue
			}

			foundPubkeys[pubkey.String()] = true
		}

		for _, pubkey := range g.pubkeys {
			if !foundPubkeys[pubkey.String()] {
				g.log.WithField("pubkey", pubkey.String()).WithField("source", node.Name()).Warn("Validator not found")

				g.metrics.UpdateStatus(MetricsStatusUnknown, []string{g.name, pubkey.String(), node.Name()})
			}
		}
	}
}

func (g *Group) checkBeaconchain(ctx context.Context, state *State) {
	g.log.Debug("Starting beaconchain validators update")

	g.beaconchainLastTick = time.Now()

	chunkIndex := 0

	for chunkIndex < len(g.beaconchainChunks) {
		// Process up to max requests per minute
		requestCount := 0
		for requestCount < g.beaconchain.GetMaxRequestsPerMinute() && chunkIndex < len(g.beaconchainChunks) {
			chunk := g.beaconchainChunks[chunkIndex]

			g.log.WithFields(logrus.Fields{
				"chunk":  chunkIndex,
				"length": len(chunk),
			}).Debug("Processing beaconchain validator pubkeys chunk")

			err := g.getValidatorsBeaconchain(ctx, chunk, state)
			if err != nil {
				g.log.WithError(err).WithField("pubkeys", chunk).Error("Error updating beaconchain validators")
			}

			requestCount++
			chunkIndex++
		}

		// Wait a minute after the last request before continuing
		if requestCount > 0 && chunkIndex < len(g.beaconchainChunks) {
			timer := time.NewTimer(time.Minute)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()

				return
			}
		}
	}

	g.log.Debug("Finished beaconchain validators update")
}

func (g *Group) getValidatorsBeaconchain(ctx context.Context, validators []string, state *State) error {
	if len(validators) == 0 {
		return nil
	}

	if len(validators) == 1 {
		response, err := g.beaconchain.GetValidator(ctx, validators[0])
		if err != nil {
			g.log.WithError(err).WithField("source", "beaconcha.in").WithField("pubkey", validators[0]).Error("Error getting validator")
			g.metrics.UpdateStatus(MetricsStatusUnknown, []string{g.name, validators[0], "beaconcha.in"})

			return err
		}

		g.updateValidatorBeaconchain(response, state)
	} else {
		response, err := g.beaconchain.GetValidators(ctx, validators)
		if err != nil {
			for _, pubkey := range validators {
				g.metrics.UpdateStatus(MetricsStatusUnknown, []string{g.name, pubkey, "beaconcha.in"})
			}

			g.log.WithError(err).WithField("source", "beaconcha.in").WithField("pubkeys", validators).Error("Error getting validators")

			return err
		}

		foundPubkeys := make(map[string]bool)

		for _, validator := range response {
			if validator != nil {
				g.updateValidatorBeaconchain(validator, state)

				foundPubkeys[validator.Pubkey] = true
			}
		}

		for _, pubkey := range g.pubkeys {
			if !foundPubkeys[pubkey.String()] {
				g.log.WithField("pubkey", pubkey.String()).WithField("source", "beaconcha.in").Warn("Validator not found")

				g.metrics.UpdateStatus(MetricsStatusUnknown, []string{g.name, pubkey.String(), "beaconcha.in"})
			}
		}
	}

	return nil
}

func (g *Group) updateValidatorBeaconchain(data *beaconchain.Validator, state *State) {
	if data == nil {
		return
	}

	source := "beaconcha.in"
	labels := []string{
		g.name,
		data.Pubkey,
		source,
	}

	g.metrics.UpdateBalance(float64(data.Balance), labels)

	status := BeaconchainToMetricsStatus(data.Status)

	credentialsCode, err := GetWithdrawalCredentialsCode(data.WithdrawalCredentials)
	if err != nil {
		g.log.WithError(err).WithField("credentials", data.WithdrawalCredentials).Error("Error parsing withdrawal credentials")
	}

	code := float64(0)
	if credentialsCode != nil {
		code = float64(*credentialsCode)

		//nolint:gosec // fine to convert as balance is always >= 0
		state.UpdateValidator(source, data.Pubkey, uint64(data.Balance), status, *credentialsCode)
	}

	g.metrics.UpdateCredentialsCode(code, labels)
	g.metrics.UpdateLastAttestationSlot(float64(data.LastAttestationSlot), labels)
	g.metrics.UpdateTotalWithdrawals(float64(data.TotalWithdrawals), labels)
	g.metrics.UpdateStatus(status, labels)
}

func (g *Group) updateValidatorBeaconAPI(data *v1.Validator, source string, state *State) {
	if data == nil {
		return
	}

	val := data.Validator

	if val == nil {
		return
	}

	labels := []string{
		g.name,
		val.PublicKey.String(),
		source,
	}

	g.metrics.UpdateBalance(float64(data.Balance), labels)

	status := BeaconAPIToMetricsStatus(data.Status, val.Slashed)

	credentialsCode, err := GetWithdrawalCredentialsCode(hex.EncodeToString(val.WithdrawalCredentials))
	if err != nil {
		g.log.WithError(err).WithField("credentials", val.WithdrawalCredentials).Error("Error parsing withdrawal credentials")
	}

	code := float64(0)

	if credentialsCode != nil {
		state.UpdateValidator(source, val.PublicKey.String(), uint64(data.Balance), status, *credentialsCode)

		code = float64(*credentialsCode)
	}

	g.metrics.UpdateCredentialsCode(code, labels)
	g.metrics.UpdateCredentialsCode(code, labels)
	g.metrics.UpdateStatus(status, labels)
}

func (g *Group) updateAlerts(changedPubkeys []string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	for _, pubkey := range changedPubkeys {
		if _, exists := g.balanceAlerts[pubkey]; !exists {
			g.balanceAlerts[pubkey] = alert.NewBalance(g.log, MinBalance)
		}

		if _, exists := g.statusAlerts[pubkey]; !exists {
			g.statusAlerts[pubkey] = alert.NewStatus(g.log, ExpectedStatuses)
		}

		if _, exists := g.withdrawalCredentialsAlerts[pubkey]; !exists {
			g.withdrawalCredentialsAlerts[pubkey] = alert.NewWithdrawalCredentials(g.log, WithdrawalCredentialsCodes)
		}

		balanceAlert := g.balanceAlerts[pubkey]
		statusAlert := g.statusAlerts[pubkey]
		withdrawalCredentialsAlert := g.withdrawalCredentialsAlerts[pubkey]

		balances := make([]uint64, 0, len(g.validatorState.Validators[pubkey].Sources))
		statuses := make([]string, 0, len(g.validatorState.Validators[pubkey].Sources))
		codes := make([]int64, 0, len(g.validatorState.Validators[pubkey].Sources))

		for _, source := range g.validatorState.Validators[pubkey].Sources {
			balances = append(balances, source.Balance)
			statuses = append(statuses, string(source.Status))
			codes = append(codes, source.WithdrawalCredentialsCode)
		}

		if shouldAlert, balance := balanceAlert.Update(balances); shouldAlert {
			g.log.WithField("balance", *balance).WithField("pubkey", pubkey).Warn("Alerting min balance")

			if err := g.publisher.Publish(validator.NewMinBalance(time.Now(), *balance, pubkey, g.name, g.monitor)); err != nil {
				g.log.WithError(err).WithField("pubkey", pubkey).Error("Error publishing min balance alert")
			}
		}

		if shouldAlert, alertingStatus := statusAlert.Update(statuses); shouldAlert {
			g.log.WithField("status", *alertingStatus).WithField("pubkey", pubkey).Warn("Alerting status")

			if err := g.publisher.Publish(validator.NewStatus(time.Now(), *alertingStatus, pubkey, g.name, g.monitor)); err != nil {
				g.log.WithError(err).WithField("pubkey", pubkey).WithField("status", *alertingStatus).Error("Error publishing status alert")
			}
		}

		if shouldAlert, alertingCredential := withdrawalCredentialsAlert.Update(codes); shouldAlert {
			g.log.WithField("credential", *alertingCredential).WithField("pubkey", pubkey).Warn("Alerting withdrawal credentials")

			if err := g.publisher.Publish(validator.NewWithdrawalCredentials(time.Now(), *alertingCredential, pubkey, g.name, g.monitor)); err != nil {
				g.log.WithError(err).WithField("pubkey", pubkey).WithField("credential", *alertingCredential).Error("Error publishing withdrawal credentials alert")
			}
		}
	}
}

func GetWithdrawalCredentialsCode(withdrawalCredentials string) (*int64, error) {
	if strings.HasPrefix(withdrawalCredentials, "0x") {
		i64, err := strconv.ParseInt(withdrawalCredentials[:4], 0, 64)
		if err != nil {
			return nil, err
		}

		return &i64, nil
	}

	i64, err := strconv.ParseInt(withdrawalCredentials[:2], 0, 64)
	if err != nil {
		return nil, err
	}

	return &i64, nil
}
