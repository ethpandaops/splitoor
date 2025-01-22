package notifier

import (
	"testing"

	"github.com/ethpandaops/splitoor/pkg/monitor/notifier/source"
	"github.com/stretchr/testify/assert"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid config - empty sources",
			config: &Config{
				Sources: []source.Config{},
			},
			expectError: false,
		},
		{
			name: "valid config - with sources",
			config: &Config{
				Sources: []source.Config{
					{
						Name:       "test_source",
						SourceType: source.SourceTypeTelegram,
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)
		})
	}
}
