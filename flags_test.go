package configlib_test

import (
	"flag"
	"os"
	"testing"

	"github.com/bherbruck/configlib"
)

func TestMultipleFlags(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		cliArgs  []string
		expected struct {
			Host  string
			Port  int
			Debug bool
		}
	}{
		{
			name: "long form flags",
			envVars: map[string]string{
				"API_KEY": "test-key",
			},
			cliArgs: []string{"--host", "example.com", "--port", "9000", "--debug"},
			expected: struct {
				Host  string
				Port  int
				Debug bool
			}{
				Host:  "example.com",
				Port:  9000,
				Debug: true,
			},
		},
		{
			name: "short form flags",
			envVars: map[string]string{
				"API_KEY": "test-key",
			},
			cliArgs: []string{"-H", "example.com", "-p", "9000", "-d"},
			expected: struct {
				Host  string
				Port  int
				Debug bool
			}{
				Host:  "example.com",
				Port:  9000,
				Debug: true,
			},
		},
		{
			name: "mixed short and long form",
			envVars: map[string]string{
				"API_KEY": "test-key",
			},
			cliArgs: []string{"-H", "example.com", "--port", "9000", "-d"},
			expected: struct {
				Host  string
				Port  int
				Debug bool
			}{
				Host:  "example.com",
				Port:  9000,
				Debug: true,
			},
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

			// Define config with multiple flags
			type Config struct {
				Host   string `env:"HOST" flag:"host,H" default:"localhost" desc:"Server host"`
				Port   int    `env:"PORT" flag:"port,p" default:"8080" desc:"Server port"`
				Debug  bool   `env:"DEBUG" flag:"debug,d" desc:"Enable debug mode"`
				APIKey string `env:"API_KEY" required:"true" desc:"API key"`
			}

			// Parse config
			var cfg Config
			err := configlib.Parse(&cfg)

			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			// Check values
			if cfg.Host != tt.expected.Host {
				t.Errorf("Host = %s, want %s", cfg.Host, tt.expected.Host)
			}
			if cfg.Port != tt.expected.Port {
				t.Errorf("Port = %d, want %d", cfg.Port, tt.expected.Port)
			}
			if cfg.Debug != tt.expected.Debug {
				t.Errorf("Debug = %v, want %v", cfg.Debug, tt.expected.Debug)
			}
		})
	}
}

func TestSliceConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		cliArgs  []string
		expected SliceConfig
	}{
		{
			name:    "default slice values",
			envVars: map[string]string{},
			cliArgs: []string{},
			expected: SliceConfig{
				Hosts: []string{"localhost", "127.0.0.1"},
				Ports: nil,
			},
		},
		{
			name: "env var slice",
			envVars: map[string]string{
				"HOSTS": "host1,host2,host3",
				"PORTS": "8080,8081,8082",
			},
			cliArgs: []string{},
			expected: SliceConfig{
				Hosts: []string{"host1", "host2", "host3"},
				Ports: []string{"8080", "8081", "8082"},
			},
		},
		{
			name: "cli override slice",
			envVars: map[string]string{
				"HOSTS": "env1,env2",
			},
			cliArgs: []string{"--hosts", "cli1,cli2,cli3"},
			expected: SliceConfig{
				Hosts: []string{"cli1", "cli2", "cli3"},
				Ports: nil,
			},
		},
		{
			name: "slice with spaces",
			envVars: map[string]string{
				"HOSTS": "host1, host2 , host3",
			},
			cliArgs: []string{},
			expected: SliceConfig{
				Hosts: []string{"host1", "host2", "host3"},
				Ports: nil,
			},
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
			resetFlagCommandLine()

			// Parse config
			var cfg SliceConfig
			err := configlib.Parse(&cfg)

			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			// Compare slices
			if !slicesEqual(cfg.Hosts, tt.expected.Hosts) {
				t.Errorf("Hosts = %v, want %v", cfg.Hosts, tt.expected.Hosts)
			}
			if !slicesEqual(cfg.Ports, tt.expected.Ports) {
				t.Errorf("Ports = %v, want %v", cfg.Ports, tt.expected.Ports)
			}
		})
	}
}

type SliceConfig struct {
	Hosts []string `env:"HOSTS" flag:"hosts" default:"localhost,127.0.0.1"`
	Ports []string `env:"PORTS" flag:"ports"`
}

// Helper function to compare string slices
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestFloatConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		cliArgs  []string
		expected FloatConfig
	}{
		{
			name:    "default float values",
			envVars: map[string]string{},
			cliArgs: []string{},
			expected: FloatConfig{
				Rate:       1.5,
				Percentage: 0.0,
				Precision:  0.0,
			},
		},
		{
			name: "env var floats",
			envVars: map[string]string{
				"RATE":       "2.5",
				"PERCENTAGE": "85.7",
				"PRECISION":  "0.001",
			},
			cliArgs: []string{},
			expected: FloatConfig{
				Rate:       2.5,
				Percentage: 85.7,
				Precision:  0.001,
			},
		},
		{
			name: "cli override floats",
			envVars: map[string]string{
				"RATE": "2.5",
			},
			cliArgs: []string{"--rate", "3.14", "--percentage", "99.9", "-p", "0.0001"},
			expected: FloatConfig{
				Rate:       3.14,
				Percentage: 99.9,
				Precision:  0.0001,
			},
		},
		{
			name: "mixed float sources",
			envVars: map[string]string{
				"PERCENTAGE": "75.0",
			},
			cliArgs: []string{"--rate", "1.618"},
			expected: FloatConfig{
				Rate:       1.618,
				Percentage: 75.0,
				Precision:  0.0,
			},
		},
		{
			name: "negative floats",
			envVars: map[string]string{
				"RATE": "-1.5",
			},
			cliArgs: []string{"--percentage", "-10.5"},
			expected: FloatConfig{
				Rate:       -1.5,
				Percentage: -10.5,
				Precision:  0.0,
			},
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
			resetFlagCommandLine()

			// Parse config
			var cfg FloatConfig
			err := configlib.Parse(&cfg)

			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			// Check values with tolerance for floating point comparison
			const tolerance = 1e-9
			if abs(cfg.Rate-tt.expected.Rate) > tolerance {
				t.Errorf("Rate = %f, want %f", cfg.Rate, tt.expected.Rate)
			}
			if abs(float64(cfg.Percentage)-float64(tt.expected.Percentage)) > tolerance {
				t.Errorf("Percentage = %f, want %f", cfg.Percentage, tt.expected.Percentage)
			}
			if abs(cfg.Precision-tt.expected.Precision) > tolerance {
				t.Errorf("Precision = %f, want %f", cfg.Precision, tt.expected.Precision)
			}
		})
	}
}

func TestFloatValidation(t *testing.T) {
	tests := []struct {
		name        string
		cliArgs     []string
		expectError bool
	}{
		{
			name:        "valid float",
			cliArgs:     []string{"--rate", "3.14"},
			expectError: false,
		},
		{
			name:        "invalid float",
			cliArgs:     []string{"--rate", "not-a-number"},
			expectError: true,
		},
		{
			name:        "scientific notation",
			cliArgs:     []string{"--rate", "1.23e-4"},
			expectError: false,
		},
		{
			name:        "integer as float",
			cliArgs:     []string{"--rate", "42"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set CLI args
			oldArgs := os.Args
			os.Args = append([]string{"test"}, tt.cliArgs...)
			defer func() { os.Args = oldArgs }()

			// Reset flag.CommandLine
			resetFlagCommandLine()

			// Parse config
			var cfg FloatConfig
			err := configlib.Parse(&cfg)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

type FloatConfig struct {
	Rate       float64 `env:"RATE" flag:"rate,r" default:"1.5" desc:"Processing rate"`
	Percentage float32 `env:"PERCENTAGE" flag:"percentage" desc:"Success percentage"`
	Precision  float64 `env:"PRECISION" flag:"precision,p" desc:"Calculation precision"`
}

// Helper function for absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Helper function to reset flag.CommandLine
func resetFlagCommandLine() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}
