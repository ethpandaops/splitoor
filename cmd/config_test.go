package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ethpandaops/splitoor/pkg/monitor/notifier/source/smtp"
	"github.com/ethpandaops/splitoor/pkg/monitor/notifier/source/telegram"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

const awkwardValue = "a\tb #c\nd: e\"f'g\\h"

func unsetEnv(t *testing.T, name string) {
	t.Helper()

	t.Setenv(name, "")
	require.NoError(t, os.Unsetenv(name))
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	return path
}

func TestDecodeConfig(t *testing.T) {
	tests := []struct {
		name    string
		content string
		set     map[string]string
		unset   []string
		want    map[string]any
		wantErr string
	}{
		{
			name:    "variable is replaced",
			content: "endpoint: https://${SAFE_HOST}/api",
			set:     map[string]string{"SAFE_HOST": "safe.example.com"},
			want:    map[string]any{"endpoint": "https://safe.example.com/api"},
		},
		{
			name:    "multiple variables in one value",
			content: "addr: ${HOST}:${PORT}",
			set:     map[string]string{"HOST": "example.com", "PORT": "8080"},
			want:    map[string]any{"addr": "example.com:8080"},
		},
		{
			name:    "default used when variable unset",
			content: "addr: ${HOST:-localhost}",
			unset:   []string{"HOST"},
			want:    map[string]any{"addr": "localhost"},
		},
		{
			name:    "default used when variable empty",
			content: "addr: ${HOST:-localhost}",
			set:     map[string]string{"HOST": ""},
			want:    map[string]any{"addr": "localhost"},
		},
		{
			name:    "default ignored when variable set",
			content: "addr: ${HOST:-localhost}",
			set:     map[string]string{"HOST": "example.com"},
			want:    map[string]any{"addr": "example.com"},
		},
		{
			name:    "empty default allows unset variable",
			content: "token: \"${TOKEN:-}\"",
			unset:   []string{"TOKEN"},
			want:    map[string]any{"token": ""},
		},
		{
			name:    "empty variable without default is an error",
			content: "token: ${TOKEN}",
			set:     map[string]string{"TOKEN": ""},
			wantErr: "config references unset or empty environment variables: TOKEN",
		},
		{
			name:    "unset variables without default are an error",
			content: "user: ${SAFE_USER}\nnested:\n  - pass: ${SAFE_API_KEY}\nagain: ${SAFE_USER}",
			unset:   []string{"SAFE_USER", "SAFE_API_KEY"},
			wantErr: "config references unset or empty environment variables: SAFE_USER, SAFE_API_KEY",
		},
		{
			name:    "comments are not expanded",
			content: "# apiKey: ${SAFE_API_KEY}\naddr: ${HOST} # ${SAFE_API_KEY}",
			set:     map[string]string{"HOST": "example.com"},
			unset:   []string{"SAFE_API_KEY"},
			want:    map[string]any{"addr": "example.com"},
		},
		{
			name:    "keys are not expanded",
			content: "${HOST}: value",
			set:     map[string]string{"HOST": "example.com"},
			want:    map[string]any{"${HOST}": "value"},
		},
		{
			name:    "literal dollar signs are left unchanged",
			content: "password: pa$$word $HOST $5 ${HOST",
			set:     map[string]string{"HOST": "example.com"},
			want:    map[string]any{"password": "pa$$word $HOST $5 ${HOST"},
		},
		{
			name:    "value with YAML syntax arrives unchanged",
			content: "plain: ${P}\ndouble: \"${P}\"\nsingle: '${P}'\nlist:\n  - ${P}\nanchored: &a ${P}\naliased: *a",
			set:     map[string]string{"P": awkwardValue},
			want: map[string]any{
				"plain":    awkwardValue,
				"double":   awkwardValue,
				"single":   awkwardValue,
				"list":     []any{awkwardValue},
				"anchored": awkwardValue,
				"aliased":  awkwardValue,
			},
		},
		{
			name:    "unquoted value resolves its type, quoted value stays a string",
			content: "port: ${PORT}\nquoted: \"${PORT}\"\nenabled: ${ENABLED}",
			set:     map[string]string{"PORT": "8080", "ENABLED": "true"},
			want:    map[string]any{"port": 8080, "quoted": "8080", "enabled": true},
		},
		{
			name:    "unquoted value that resolves to null stays a string",
			content: "name: ${NAME}\ntilde: ${TILDE}\nempty: ${EMPTY:-}",
			set:     map[string]string{"NAME": "null", "TILDE": "~"},
			unset:   []string{"EMPTY"},
			want:    map[string]any{"name": "null", "tilde": "~", "empty": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.set {
				t.Setenv(k, v)
			}

			for _, k := range tt.unset {
				unsetEnv(t, k)
			}

			var got map[string]any

			err := decodeConfig([]byte(tt.content), &got)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDecodeConfigWithoutPlaceholdersIsUnchanged(t *testing.T) {
	data, err := os.ReadFile("../example_config.yaml")
	require.NoError(t, err)

	var want, got any

	require.NoError(t, yaml.Unmarshal(data, &want))
	require.NoError(t, decodeConfig(data, &got))
	assert.Equal(t, want, got)
}

func TestLoadMonitorConfigExpandsEnvVars(t *testing.T) {
	t.Setenv("SPLITOOR_TEST_SAFE_API_KEY", awkwardValue)
	t.Setenv("SPLITOOR_TEST_AUTH", awkwardValue)
	t.Setenv("SPLITOOR_TEST_BATCH_SIZE", "25")
	t.Setenv("SPLITOOR_TEST_BOT_TOKEN", awkwardValue)
	t.Setenv("SPLITOOR_TEST_SMTP_PORT", "2525")

	config, err := loadMonitorConfigFromFile(writeConfig(t, `
name: monitor-1
ethereum:
  execution:
    - name: execution-1
      nodeAddress: http://localhost:8545
      nodeHeaders:
        Authorization: "${SPLITOOR_TEST_AUTH}"
safe:
  apiKey: ${SPLITOOR_TEST_SAFE_API_KEY}
beaconchain:
  batchSize: ${SPLITOOR_TEST_BATCH_SIZE}
notifier:
  sources:
    - type: telegram
      name: telegram
      config:
        botToken: ${SPLITOOR_TEST_BOT_TOKEN}
        chatId: "-1001234"
    - type: smtp
      name: smtp
      config:
        port: ${SPLITOOR_TEST_SMTP_PORT}
`))
	require.NoError(t, err)

	assert.Equal(t, "monitor-1", config.Name)
	assert.Equal(t, awkwardValue, config.Safe.APIKey)
	assert.Equal(t, 25, config.Beaconchain.BatchSize)
	require.Len(t, config.Ethereum.Execution, 1)
	assert.Equal(t, awkwardValue, config.Ethereum.Execution[0].NodeHeaders["Authorization"])
	require.Len(t, config.Notifier.Sources, 2)

	telegramConfig := &telegram.Config{}

	require.NoError(t, config.Notifier.Sources[0].Config.Unmarshal(telegramConfig))
	assert.Equal(t, awkwardValue, telegramConfig.BotToken)
	assert.Equal(t, "-1001234", telegramConfig.ChatID)

	smtpConfig := &smtp.Config{}

	require.NoError(t, config.Notifier.Sources[1].Config.Unmarshal(smtpConfig))
	assert.Equal(t, 2525, smtpConfig.Port)
}

func TestLoadMonitorConfigFailsOnUnsetEnvVar(t *testing.T) {
	unsetEnv(t, "SPLITOOR_TEST_SAFE_API_KEY")

	_, err := loadMonitorConfigFromFile(writeConfig(t, `
safe:
  apiKey: "${SPLITOOR_TEST_SAFE_API_KEY}"
`))
	require.ErrorContains(t, err, "SPLITOOR_TEST_SAFE_API_KEY")
}
