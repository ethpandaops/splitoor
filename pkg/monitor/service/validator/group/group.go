package group

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/ethpandaops/splitoor/pkg/ethereum"
	"github.com/ethpandaops/splitoor/pkg/monitor/beaconchain"
	"github.com/sirupsen/logrus"
)

type Group struct {
	log  logrus.FieldLogger
	name string

	ethereumPool *ethereum.Pool
	pubkeys      []phase0.BLSPubKey

	beaconchain         beaconchain.Client
	beaconchainChunks   [][]string
	beaconchainLastTick time.Time

	metrics *Metrics
}

func NewGroup(ctx context.Context, log logrus.FieldLogger, monitor string, conf *Config, ethereumPool *ethereum.Pool, bc beaconchain.Client) (*Group, error) {
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

	return &Group{
		log:               log.WithField("module", "group"),
		name:              conf.Name,
		ethereumPool:      ethereumPool,
		pubkeys:           pubkeys,
		beaconchain:       bc,
		beaconchainChunks: chunks,
		metrics:           GetMetricsInstance("splitoor_validator", monitor),
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

func (g *Group) tick(ctx context.Context) {
	if g.beaconchain != nil && time.Since(g.beaconchainLastTick) > g.beaconchain.GetCheckInterval() {
		go g.tickBeaconchain(ctx)
	}

	if g.ethereumPool != nil && g.ethereumPool.HasHealthyBeaconNodes() {
		go g.tickBeaconAPI(ctx)
	}
}

func (g *Group) tickBeaconAPI(ctx context.Context) {
	for _, node := range g.ethereumPool.GetHealthyBeaconNodes() {
		validators, err := node.Node().FetchValidators(ctx, "head", nil, g.pubkeys)
		if err != nil {
			g.log.WithError(err).WithField("node", node.Name()).Error("Error fetching validators")
		}

		for _, validator := range validators {
			g.updateValidatorMetricsBeaconAPI(validator, node.Name())
		}
	}
}

func (g *Group) tickBeaconchain(ctx context.Context) {
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

			err := g.getValidatorsBeaconchain(ctx, chunk)
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

func (g *Group) getValidatorsBeaconchain(ctx context.Context, validators []string) error {
	if len(validators) == 0 {
		return nil
	}

	if len(validators) == 1 {
		response, err := g.beaconchain.GetValidator(ctx, validators[0])
		if err != nil {
			return err
		}

		g.updateValidatorMetricsBeaconchain(response)
	} else {
		response, err := g.beaconchain.GetValidators(ctx, validators)
		if err != nil {
			return err
		}

		for _, validator := range response {
			if validator != nil {
				g.updateValidatorMetricsBeaconchain(validator)
			}
		}
	}

	return nil
}

func (g *Group) updateValidatorMetricsBeaconchain(data *beaconchain.Validator) {
	if data == nil {
		return
	}

	labels := []string{
		g.name,
		data.Pubkey,
		"beaconcha.in",
	}

	g.metrics.UpdateBalance(float64(data.Balance), labels)

	credentialsCode, err := GetWithdrawalCredentialsCode(data.WithdrawalCredentials)
	if err != nil {
		g.log.WithError(err).WithField("credentials", data.WithdrawalCredentials).Error("Error parsing withdrawal credentials")
	}

	code := float64(0)
	if credentialsCode != nil {
		code = float64(*credentialsCode)
	}

	g.metrics.UpdateCredentialsCode(code, labels)
	g.metrics.UpdateLastAttestationSlot(float64(data.LastAttestationSlot), labels)
	g.metrics.UpdateTotalWithdrawals(float64(data.TotalWithdrawals), labels)
	g.metrics.UpdateStatus(BeaconchainToMetricsStatus(data.Status), labels)
}

func (g *Group) updateValidatorMetricsBeaconAPI(data *v1.Validator, source string) {
	if data == nil {
		return
	}

	validator := data.Validator

	if validator == nil {
		return
	}

	labels := []string{
		g.name,
		validator.PublicKey.String(),
		source,
	}

	g.metrics.UpdateBalance(float64(data.Balance), labels)

	credentialsCode, err := GetWithdrawalCredentialsCode(hex.EncodeToString(validator.WithdrawalCredentials))
	if err != nil {
		g.log.WithError(err).WithField("credentials", validator.WithdrawalCredentials).Error("Error parsing withdrawal credentials")
	}

	code := float64(0)
	if credentialsCode != nil {
		code = float64(*credentialsCode)
	}

	g.metrics.UpdateCredentialsCode(code, labels)
	g.metrics.UpdateStatus(BeaconAPIToMetricsStatus(data.Status, validator.Slashed), labels)
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
