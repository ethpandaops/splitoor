package ses_test

import (
	"testing"

	"github.com/ethpandaops/splitoor/pkg/monitor/notifier/source/ses"
	"github.com/stretchr/testify/assert"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *ses.Config
		expectError bool
	}{
		{
			name: "valid config",
			config: &ses.Config{
				From: "test@example.com",
				To:   []string{"recipient@example.com"},
			},
			expectError: false,
		},
		{
			name: "empty from",
			config: &ses.Config{
				To: []string{"recipient@example.com"},
			},
			expectError: true,
		},
		{
			name: "empty recipients",
			config: &ses.Config{
				From: "test@example.com",
			},
			expectError: true,
		},
		{
			name: "empty recipient list",
			config: &ses.Config{
				From: "test@example.com",
				To:   []string{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
