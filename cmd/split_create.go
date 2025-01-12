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
	createElRPCURL           string
	createDeployerAddress    string
	createDeployerPrivateKey string
	createContractAddress    string
	createRecipients         string
	createPercentages        string
	createController         string
	createDistributorFee     uint32
	createGasLimit           uint64
)

var createSplitCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new split",
	Long:  `Create a new split with specified recipients and percentages.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		initCommon()

		err := createSplit(cmd.Context())
		if err != nil {
			log.Fatal(err)
		}

		return nil
	},
}

func init() {
	splitCmd.AddCommand(createSplitCmd)

	createSplitCmd.Flags().StringVar(&createElRPCURL, "el-rpc-url", "", "Execution layer (EL) RPC URL")
	createSplitCmd.Flags().StringVar(&createDeployerAddress, "deployer-address", "", "Deployer address")
	createSplitCmd.Flags().StringVar(&createDeployerPrivateKey, "deployer-private-key", "", "Deployer private key")
	createSplitCmd.Flags().StringVar(&createContractAddress, "contract", "", "Contract address")
	createSplitCmd.Flags().StringVar(&createRecipients, "recipients", "", "Comma-separated list of recipient addresses")
	createSplitCmd.Flags().StringVar(&createPercentages, "percentages", "", "Comma-separated list of percentages as an integer where 999999 = 99.9999%. Must sum to 1000000")
	createSplitCmd.Flags().StringVar(&createController, "controller", "", "Controller address")
	createSplitCmd.Flags().Uint32Var(&createDistributorFee, "distributor-fee", 0, "Distributor fee percentage")
	createSplitCmd.Flags().Uint64Var(&createGasLimit, "gaslimit", 3000000, "Gas limit for transaction")

	err := createSplitCmd.MarkFlagRequired("el-rpc-url")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "el-rpc-url")
	}

	err = createSplitCmd.MarkFlagRequired("deployer-address")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "deployer-address")
	}

	err = createSplitCmd.MarkFlagRequired("deployer-private-key")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "deployer-private-key")
	}

	err = createSplitCmd.MarkFlagRequired("recipients")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "recipients")
	}

	err = createSplitCmd.MarkFlagRequired("percentages")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "percentages")
	}

	err = createSplitCmd.MarkFlagRequired("controller")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "controller")
	}
}

func createSplit(ctx context.Context) error {
	dpNodeConfig := &execution.Config{
		NodeAddress: createElRPCURL,
	}

	dpNode := execution.NewNode(log, "execution", dpNodeConfig)
	if err := dpNode.Start(ctx); err != nil {
		return errors.Wrap(err, "failed to start execution node")
	}

	if createContractAddress == "" {
		address, err := getSplitDefaultContractAddress(ctx, dpNode)
		if err != nil {
			return err
		}

		createContractAddress = *address
	}

	config := &split.Config{
		ContractAddress: createContractAddress,
	}

	client, err := split.NewClient(log, config)
	if err != nil {
		return err
	}

	contractABI, err := contract.GetSplitMainAbi()
	if err != nil {
		return err
	}

	accounts, allocations, err := parseRecipients(createRecipients, createPercentages)
	if err != nil {
		return err
	}

	params := split.CreateSplitParams{
		Controller:            createController,
		Accounts:              accounts,
		PercentageAllocations: allocations,
		DistributorFee:        createDistributorFee,
	}

	splitAddress, err := client.Create(ctx, dpNode, contractABI, createDeployerAddress, createDeployerPrivateKey, createGasLimit, &params)
	if err != nil {
		return err
	}

	log.WithFields(logrus.Fields{
		"split": *splitAddress,
	}).Info("Split created")

	return nil
}
