package cmd

import (
	"context"

	"github.com/ethpandaops/splitoor/pkg/0xsplits/contract"
	"github.com/ethpandaops/splitoor/pkg/0xsplits/split"
	"github.com/ethpandaops/splitoor/pkg/ethereum/execution"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	distributeElRPCURL           string
	distributeDeployerAddress    string
	distributeDeployerPrivKey    string
	distributeContractAddress    string
	distributeSplitAddress       string
	distributeRecipients         string
	distributePercentages        string
	distributeDistributorFee     uint32
	distributeDistributorAddress string
	distributeGasLimit           uint64
)

var distributeSplitCmd = &cobra.Command{
	Use:   "distribute",
	Short: "Distribute an existing split",
	Long:  `Distribute an existing split with new recipients and percentages.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		initCommon()

		err := distributeSplit(cmd.Context())
		if err != nil {
			log.Fatal(err)
		}

		return nil
	},
}

func init() {
	splitCmd.AddCommand(distributeSplitCmd)

	distributeSplitCmd.Flags().StringVar(&distributeElRPCURL, "el-rpc-url", "", "Execution layer (EL) RPC URL")
	distributeSplitCmd.Flags().StringVar(&distributeDeployerAddress, "deployer-address", "", "Deployer address")
	distributeSplitCmd.Flags().StringVar(&distributeDeployerPrivKey, "deployer-private-key", "", "Deployer private key")
	distributeSplitCmd.Flags().StringVar(&distributeContractAddress, "contract", "", "Contract address")
	distributeSplitCmd.Flags().StringVar(&distributeSplitAddress, "split", "", "Split address to distribute")
	distributeSplitCmd.Flags().StringVar(&distributeRecipients, "recipients", "", "Comma-separated list of recipient addresses")
	distributeSplitCmd.Flags().StringVar(&distributePercentages, "percentages", "", "Comma-separated list of percentages as an integer where 999999 = 99.9999%. Must sum to 1000000")
	distributeSplitCmd.Flags().Uint32Var(&distributeDistributorFee, "distributor-fee", 0, "Distributor fee percentage")
	distributeSplitCmd.Flags().StringVar(&distributeDistributorAddress, "distributor-address", "", "Distributor address")
	distributeSplitCmd.Flags().Uint64Var(&distributeGasLimit, "gaslimit", 3000000, "Gas limit for transaction")

	err := distributeSplitCmd.MarkFlagRequired("el-rpc-url")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "el-rpc-url")
	}

	err = distributeSplitCmd.MarkFlagRequired("deployer-address")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "deployer-address")
	}

	err = distributeSplitCmd.MarkFlagRequired("deployer-private-key")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "deployer-private-key")
	}

	err = distributeSplitCmd.MarkFlagRequired("split")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "split")
	}

	err = distributeSplitCmd.MarkFlagRequired("recipients")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "recipients")
	}

	err = distributeSplitCmd.MarkFlagRequired("percentages")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "percentages")
	}
}

func distributeSplit(ctx context.Context) error {
	dpNodeConfig := &execution.Config{
		NodeAddress: distributeElRPCURL,
	}

	dpNode := execution.NewNode(log, "execution", dpNodeConfig)
	if err := dpNode.Start(ctx); err != nil {
		return errors.Wrap(err, "failed to start execution node")
	}

	if distributeContractAddress == "" {
		address, err := getSplitDefaultContractAddress(ctx, dpNode)
		if err != nil {
			return err
		}

		distributeContractAddress = *address
	}

	config := &split.Config{
		ContractAddress: distributeContractAddress,
		SplitAddress:    &distributeSplitAddress,
	}

	client, err := split.NewClient(log, config)
	if err != nil {
		return err
	}

	contractABI, err := contract.GetSplitMainAbi()
	if err != nil {
		return err
	}

	accounts, allocations, err := parseRecipients(distributeRecipients, distributePercentages)
	if err != nil {
		return err
	}

	params := split.DistributeETHParams{
		Accounts:              accounts,
		PercentageAllocations: allocations,
		DistributorFee:        distributeDistributorFee,
		DistributorAddress:    distributeDistributorAddress,
	}

	err = client.DistributeETH(ctx, dpNode, contractABI, distributeDeployerAddress, distributeDeployerPrivKey, distributeGasLimit, &params)
	if err != nil {
		return err
	}

	log.Info("Split distributed")

	return nil
}
