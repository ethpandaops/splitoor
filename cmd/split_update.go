package cmd

import (
	"context"

	"github.com/ethpandaops/splitoor/pkg/0xsplits/contract"
	"github.com/ethpandaops/splitoor/pkg/0xsplits/split"
	"github.com/ethpandaops/splitoor/pkg/ethereum/execution"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	updateElRPCURL        string
	updateDeployerAddress string
	updateDeployerPrivKey string
	updateContractAddress string
	updateSplitAddress    string
	updateRecipients      string
	updatePercentages     string
	updateDistributorFee  uint32
	updateGasLimit        uint64
)

var updateSplitCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing split",
	Long:  `Update an existing split with new recipients and percentages.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		initCommon()

		err := updateSplit(cmd.Context())
		if err != nil {
			log.Fatal(err)
		}

		return nil
	},
}

func init() {
	splitCmd.AddCommand(updateSplitCmd)

	updateSplitCmd.Flags().StringVar(&updateElRPCURL, "el-rpc-url", "", "Execution layer (EL) RPC URL")
	updateSplitCmd.Flags().StringVar(&updateDeployerAddress, "deployer-address", "", "Deployer address")
	updateSplitCmd.Flags().StringVar(&updateDeployerPrivKey, "deployer-private-key", "", "Deployer private key")
	updateSplitCmd.Flags().StringVar(&updateContractAddress, "contract", "", "Contract address")
	updateSplitCmd.Flags().StringVar(&updateSplitAddress, "split", "", "Split address to update")
	updateSplitCmd.Flags().StringVar(&updateRecipients, "recipients", "", "Comma-separated list of recipient addresses")
	updateSplitCmd.Flags().StringVar(&updatePercentages, "percentages", "", "Comma-separated list of percentages as an integer where 999999 = 99.9999%. Must sum to 1000000")
	updateSplitCmd.Flags().Uint32Var(&updateDistributorFee, "distributor-fee", 0, "Distributor fee percentage")
	updateSplitCmd.Flags().Uint64Var(&updateGasLimit, "gaslimit", 3000000, "Gas limit for transaction")

	err := updateSplitCmd.MarkFlagRequired("el-rpc-url")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "el-rpc-url")
	}

	err = updateSplitCmd.MarkFlagRequired("deployer-address")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "deployer-address")
	}

	err = updateSplitCmd.MarkFlagRequired("deployer-private-key")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "deployer-private-key")
	}

	err = updateSplitCmd.MarkFlagRequired("split")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "split")
	}

	err = updateSplitCmd.MarkFlagRequired("recipients")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "recipients")
	}

	err = updateSplitCmd.MarkFlagRequired("percentages")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "percentages")
	}
}

func updateSplit(ctx context.Context) error {
	dpNodeConfig := &execution.Config{
		NodeAddress: updateElRPCURL,
	}

	dpNode := execution.NewNode(log, "execution", dpNodeConfig)
	if err := dpNode.Start(ctx); err != nil {
		return errors.Wrap(err, "failed to start execution node")
	}

	if updateContractAddress == "" {
		address, err := getSplitDefaultContractAddress(ctx, dpNode)
		if err != nil {
			return err
		}

		updateContractAddress = *address
	}

	config := &split.Config{
		ContractAddress: updateContractAddress,
		SplitAddress:    &updateSplitAddress,
	}

	client, err := split.NewClient(log, config)
	if err != nil {
		return err
	}

	contractABI, err := contract.GetSplitMainAbi()
	if err != nil {
		return err
	}

	accounts, allocations, err := parseRecipients(updateRecipients, updatePercentages)
	if err != nil {
		return err
	}

	params := split.UpdateSplitParams{
		Accounts:              accounts,
		PercentageAllocations: allocations,
		DistributorFee:        updateDistributorFee,
	}

	splitAddress, err := client.Update(ctx, dpNode, contractABI, updateDeployerAddress, updateDeployerPrivKey, updateGasLimit, &params)
	if err != nil {
		return err
	}

	log.WithFields(logrus.Fields{
		"split": *splitAddress,
	}).Info("Split updated")

	return nil
}
