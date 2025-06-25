package configlib_test

import (
	"flag"
	"os"
	"testing"

	"github.com/bherbruck/configlib"
)

// Test structures used across multiple test files
type SimpleConfig struct {
	Host     string `env:"HOST" cli:"host" default:"localhost" desc:"Server host"`
	Port     int    `env:"PORT" cli:"port" default:"8080" desc:"Server port"`
	Debug    bool   `env:"DEBUG" cli:"debug" default:"false" desc:"Enable debug mode"`
	Required string `env:"REQUIRED" cli:"required" required:"true" desc:"Required field"`
}

type NestedConfig struct {
	Server struct {
		Host string `env:"SERVER_HOST" cli:"server-host" required:"true"`
		Port int    `env:"SERVER_PORT" cli:"server-port" default:"8080"`
	}
	Database struct {
		Host     string `env:"DB_HOST" cli:"db-host" required:"true"`
		Port     int    `env:"DB_PORT" cli:"db-port" default:"5432"`
		Password string `env:"DB_PASSWORD" cli:"db-password" required:"true"`
	}
}

func TestSimpleConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		cliArgs  []string
		expected SimpleConfig
		wantErr  bool
	}{
		{
			name: "all defaults with required field",
			envVars: map[string]string{
				"REQUIRED": "test-value",
			},
			cliArgs: []string{},
			expected: SimpleConfig{
				Host:     "localhost",
				Port:     8080,
				Debug:    false,
				Required: "test-value",
			},
			wantErr: false,
		},
		{
			name:    "missing required field",
			envVars: map[string]string{},
			cliArgs: []string{},
			wantErr: true,
		},
		{
			name: "env vars override defaults",
			envVars: map[string]string{
				"HOST":     "example.com",
				"PORT":     "9000",
				"DEBUG":    "true",
				"REQUIRED": "test",
			},
			cliArgs: []string{},
			expected: SimpleConfig{
				Host:     "example.com",
				Port:     9000,
				Debug:    true,
				Required: "test",
			},
			wantErr: false,
		},
		{
			name: "cli args override env vars",
			envVars: map[string]string{
				"HOST":     "example.com",
				"PORT":     "9000",
				"REQUIRED": "test",
			},
			cliArgs: []string{"--host", "cli-host", "--port", "3000"},
			expected: SimpleConfig{
				Host:     "cli-host",
				Port:     3000,
				Debug:    false,
				Required: "test",
			},
			wantErr: false,
		},
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

			// Reset flag.CommandLine to avoid conflicts
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			// Parse config
			var cfg SimpleConfig
			err := configlib.Parse(&cfg)

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && cfg != tt.expected {
				t.Errorf("Parse() got = %+v, want %+v", cfg, tt.expected)
			}
		})
	}
}

func TestNestedConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		cliArgs  []string
		validate func(*testing.T, *NestedConfig)
		wantErr  bool
	}{
		{
			name: "all required fields provided",
			envVars: map[string]string{
				"SERVER_HOST": "localhost",
				"DB_HOST":     "db.example.com",
				"DB_PASSWORD": "secret",
			},
			validate: func(t *testing.T, cfg *NestedConfig) {
				if cfg.Server.Host != "localhost" {
					t.Errorf("Server.Host = %s, want localhost", cfg.Server.Host)
				}
				if cfg.Server.Port != 8080 {
					t.Errorf("Server.Port = %d, want 8080", cfg.Server.Port)
				}
				if cfg.Database.Host != "db.example.com" {
					t.Errorf("Database.Host = %s, want db.example.com", cfg.Database.Host)
				}
				if cfg.Database.Port != 5432 {
					t.Errorf("Database.Port = %d, want 5432", cfg.Database.Port)
				}
				if cfg.Database.Password != "secret" {
					t.Errorf("Database.Password = %s, want secret", cfg.Database.Password)
				}
			},
			wantErr: false,
		},
		{
			name: "cli overrides env",
			envVars: map[string]string{
				"SERVER_HOST": "env-host",
				"DB_HOST":     "env-db",
				"DB_PASSWORD": "env-pass",
			},
			cliArgs: []string{"--server-host", "cli-host", "--db-password", "cli-pass"},
			validate: func(t *testing.T, cfg *NestedConfig) {
				if cfg.Server.Host != "cli-host" {
					t.Errorf("Server.Host = %s, want cli-host", cfg.Server.Host)
				}
				if cfg.Database.Password != "cli-pass" {
					t.Errorf("Database.Password = %s, want cli-pass", cfg.Database.Password)
				}
			},
			wantErr: false,
		},
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
			var cfg NestedConfig
			err := configlib.Parse(&cfg)

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, &cfg)
			}
		})
	}
}
