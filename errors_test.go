package configlib_test

import (
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/bherbruck/configlib"
)

func TestMultipleMissingRequiredFields(t *testing.T) {
	// Clear environment
	os.Clearenv()

	// Reset CLI args
	oldArgs := os.Args
	os.Args = []string{"test"}
	defer func() { os.Args = oldArgs }()

	// Reset flag.CommandLine
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	var cfg NestedConfig
	err := configlib.Parse(&cfg)

	if err == nil {
		t.Fatal("Expected error for missing required fields, got nil")
	}

	errMsg := err.Error()

	// Check that the error message contains all missing fields
	expectedFields := []string{
		"Server.Host (env: SERVER_HOST, cli: --server-host)",
		"Database.Host (env: DB_HOST, cli: --db-host)",
		"Database.Password (env: DB_PASSWORD, cli: --db-password)",
	}

	if !strings.Contains(errMsg, "missing required fields:") {
		t.Errorf("Error message should start with 'missing required fields:', got: %s", errMsg)
	}

	for _, field := range expectedFields {
		if !strings.Contains(errMsg, field) {
			t.Errorf("Error message should contain '%s', got: %s", field, errMsg)
		}
	}
}

func TestInvalidValues(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		cliArgs []string
		wantErr bool
		errMsg  string
	}{
		{
			name: "invalid int from env",
			envVars: map[string]string{
				"PORT":     "not-a-number",
				"REQUIRED": "test",
			},
			wantErr: true,
			errMsg:  "error setting field Port",
		},
		{
			name: "invalid bool from env",
			envVars: map[string]string{
				"DEBUG":    "not-a-bool",
				"REQUIRED": "test",
			},
			wantErr: true,
			errMsg:  "error setting field Debug",
		},
		{
			name: "invalid int from cli",
			envVars: map[string]string{
				"REQUIRED": "test",
			},
			cliArgs: []string{"--port", "not-a-number"},
			wantErr: true,
			errMsg:  "invalid integer value",
		},
		// Note: We don't test invalid bool from CLI because boolean flags
		// using BoolVar don't take explicit values - they're either present (true)
		// or absent (false). "--debug not-a-bool" is parsed as --debug (true)
		// followed by a non-flag argument "not-a-bool".
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Set CLI args
			oldArgs := os.Args
			os.Args = append([]string{"test"}, tt.cliArgs...)
			defer func() { os.Args = oldArgs }()

			// Reset flag.CommandLine
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			// Parse config
			var cfg SimpleConfig
			err := configlib.Parse(&cfg)

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Error message should contain '%s', got: %s", tt.errMsg, err.Error())
			}
		})
	}
}
