package cmd

import (
	"context"
	"os"

	"github.com/creasty/defaults"
	m "github.com/ethpandaops/splitoor/pkg/monitor"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var monitorConfigFile string

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor",
	Long:  `Monitor splits`,
	RunE: func(cmd *cobra.Command, args []string) error {
		initCommon()

		monitor(cmd.Context())

		return nil
	},
}

func init() {
	rootCmd.AddCommand(monitorCmd)

	monitorCmd.Flags().StringVar(&monitorConfigFile, "config", "config.yaml", "Config file (default is config.yaml)")
}

func monitor(ctx context.Context) {
	config, err := loadMonitorConfigFromFile(monitorConfigFile)
	if err != nil {
		log.Fatal(err)
	}

	server, err := m.NewServer(ctx, log, config)
	if err != nil {
		log.Fatal(err)
	}

	if startErr := server.Start(ctx); startErr != nil {
		log.Fatal(startErr)
	}

	log.Info("Splitoor monitor exited - cya!")
}

func loadMonitorConfigFromFile(file string) (*m.Config, error) {
	if file == "" {
		file = "config.yaml"
	}

	config := &m.Config{}

	if defaultsErr := defaults.Set(config); defaultsErr != nil {
		return nil, defaultsErr
	}

	yamlFile, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	type plain m.Config

	if unmarshalErr := yaml.Unmarshal(yamlFile, (*plain)(config)); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return config, nil
}
