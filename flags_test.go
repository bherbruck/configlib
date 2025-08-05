package configlib_test

import (
	"flag"
	"os"
	"testing"
	"time"

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

func TestIntegerTypesConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		cliArgs  []string
		expected IntegerTypesConfig
	}{
		{
			name:    "default integer values",
			envVars: map[string]string{},
			cliArgs: []string{},
			expected: IntegerTypesConfig{
				Int8Val:   42,
				Int16Val:  1000,
				Int32Val:  100000,
				Int64Val:  9223372036854775807,
				UintVal:   100,
				Uint8Val:  255,
				Uint16Val: 65535,
				Uint32Val: 4294967295,
				Uint64Val: 18446744073709551615,
			},
		},
		{
			name: "env var integers",
			envVars: map[string]string{
				"INT8_VAL":   "-128",
				"INT16_VAL":  "-32768",
				"INT32_VAL":  "-2147483648",
				"INT64_VAL":  "-9223372036854775808",
				"UINT_VAL":   "200",
				"UINT8_VAL":  "128",
				"UINT16_VAL": "32768",
				"UINT32_VAL": "2147483648",
				"UINT64_VAL": "9223372036854775808",
			},
			cliArgs: []string{},
			expected: IntegerTypesConfig{
				Int8Val:   -128,
				Int16Val:  -32768,
				Int32Val:  -2147483648,
				Int64Val:  -9223372036854775808,
				UintVal:   200,
				Uint8Val:  128,
				Uint16Val: 32768,
				Uint32Val: 2147483648,
				Uint64Val: 9223372036854775808,
			},
		},
		{
			name: "cli override integers",
			envVars: map[string]string{
				"INT8_VAL": "10",
			},
			cliArgs: []string{"--int8", "20", "--uint64", "999"},
			expected: IntegerTypesConfig{
				Int8Val:   20,
				Int16Val:  1000,
				Int32Val:  100000,
				Int64Val:  9223372036854775807,
				UintVal:   100,
				Uint8Val:  255,
				Uint16Val: 65535,
				Uint32Val: 4294967295,
				Uint64Val: 999,
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
			var cfg IntegerTypesConfig
			err := configlib.Parse(&cfg)

			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			// Check values
			if cfg.Int8Val != tt.expected.Int8Val {
				t.Errorf("Int8Val = %d, want %d", cfg.Int8Val, tt.expected.Int8Val)
			}
			if cfg.Int16Val != tt.expected.Int16Val {
				t.Errorf("Int16Val = %d, want %d", cfg.Int16Val, tt.expected.Int16Val)
			}
			if cfg.Int32Val != tt.expected.Int32Val {
				t.Errorf("Int32Val = %d, want %d", cfg.Int32Val, tt.expected.Int32Val)
			}
			if cfg.Int64Val != tt.expected.Int64Val {
				t.Errorf("Int64Val = %d, want %d", cfg.Int64Val, tt.expected.Int64Val)
			}
			if cfg.UintVal != tt.expected.UintVal {
				t.Errorf("UintVal = %d, want %d", cfg.UintVal, tt.expected.UintVal)
			}
			if cfg.Uint8Val != tt.expected.Uint8Val {
				t.Errorf("Uint8Val = %d, want %d", cfg.Uint8Val, tt.expected.Uint8Val)
			}
			if cfg.Uint16Val != tt.expected.Uint16Val {
				t.Errorf("Uint16Val = %d, want %d", cfg.Uint16Val, tt.expected.Uint16Val)
			}
			if cfg.Uint32Val != tt.expected.Uint32Val {
				t.Errorf("Uint32Val = %d, want %d", cfg.Uint32Val, tt.expected.Uint32Val)
			}
			if cfg.Uint64Val != tt.expected.Uint64Val {
				t.Errorf("Uint64Val = %d, want %d", cfg.Uint64Val, tt.expected.Uint64Val)
			}
		})
	}
}

func TestDurationConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		cliArgs  []string
		expected DurationConfig
	}{
		{
			name:    "default duration values",
			envVars: map[string]string{},
			cliArgs: []string{},
			expected: DurationConfig{
				Timeout:  30 * time.Second,
				Interval: 5 * time.Minute,
				MaxWait:  0,
			},
		},
		{
			name: "env var durations",
			envVars: map[string]string{
				"TIMEOUT":  "1m30s",
				"INTERVAL": "10s",
				"MAX_WAIT": "2h",
			},
			cliArgs: []string{},
			expected: DurationConfig{
				Timeout:  90 * time.Second,
				Interval: 10 * time.Second,
				MaxWait:  2 * time.Hour,
			},
		},
		{
			name: "cli override durations",
			envVars: map[string]string{
				"TIMEOUT": "1m",
			},
			cliArgs: []string{"--timeout", "45s", "--max-wait", "30m"},
			expected: DurationConfig{
				Timeout:  45 * time.Second,
				Interval: 5 * time.Minute,
				MaxWait:  30 * time.Minute,
			},
		},
		{
			name: "complex duration formats",
			envVars: map[string]string{
				"TIMEOUT": "1h30m45s",
			},
			cliArgs: []string{"--interval", "2m30s"},
			expected: DurationConfig{
				Timeout:  time.Hour + 30*time.Minute + 45*time.Second,
				Interval: 2*time.Minute + 30*time.Second,
				MaxWait:  0,
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
			var cfg DurationConfig
			err := configlib.Parse(&cfg)

			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			// Check values
			if cfg.Timeout != tt.expected.Timeout {
				t.Errorf("Timeout = %v, want %v", cfg.Timeout, tt.expected.Timeout)
			}
			if cfg.Interval != tt.expected.Interval {
				t.Errorf("Interval = %v, want %v", cfg.Interval, tt.expected.Interval)
			}
			if cfg.MaxWait != tt.expected.MaxWait {
				t.Errorf("MaxWait = %v, want %v", cfg.MaxWait, tt.expected.MaxWait)
			}
		})
	}
}

func TestIntegerValidation(t *testing.T) {
	tests := []struct {
		name        string
		cliArgs     []string
		expectError bool
	}{
		{
			name:        "valid int64",
			cliArgs:     []string{"--int64", "9223372036854775807"},
			expectError: false,
		},
		{
			name:        "invalid int64 overflow",
			cliArgs:     []string{"--int64", "9223372036854775808"},
			expectError: true,
		},
		{
			name:        "valid uint64",
			cliArgs:     []string{"--uint64", "18446744073709551615"},
			expectError: false,
		},
		{
			name:        "invalid uint64 negative",
			cliArgs:     []string{"--uint64", "-1"},
			expectError: true,
		},
		{
			name:        "invalid integer format",
			cliArgs:     []string{"--int32", "not-a-number"},
			expectError: true,
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
			var cfg IntegerTypesConfig
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

func TestDurationValidation(t *testing.T) {
	tests := []struct {
		name        string
		cliArgs     []string
		expectError bool
	}{
		{
			name:        "valid duration",
			cliArgs:     []string{"--timeout", "30s"},
			expectError: false,
		},
		{
			name:        "invalid duration format",
			cliArgs:     []string{"--timeout", "not-a-duration"},
			expectError: true,
		},
		{
			name:        "complex valid duration",
			cliArgs:     []string{"--timeout", "1h30m45s"},
			expectError: false,
		},
		{
			name:        "microseconds duration",
			cliArgs:     []string{"--timeout", "500µs"},
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
			var cfg DurationConfig
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

type IntegerTypesConfig struct {
	Int8Val   int8   `env:"INT8_VAL" flag:"int8" default:"42" desc:"8-bit signed integer"`
	Int16Val  int16  `env:"INT16_VAL" flag:"int16" default:"1000" desc:"16-bit signed integer"`
	Int32Val  int32  `env:"INT32_VAL" flag:"int32" default:"100000" desc:"32-bit signed integer"`
	Int64Val  int64  `env:"INT64_VAL" flag:"int64" default:"9223372036854775807" desc:"64-bit signed integer"`
	UintVal   uint   `env:"UINT_VAL" flag:"uint" default:"100" desc:"Unsigned integer"`
	Uint8Val  uint8  `env:"UINT8_VAL" flag:"uint8" default:"255" desc:"8-bit unsigned integer"`
	Uint16Val uint16 `env:"UINT16_VAL" flag:"uint16" default:"65535" desc:"16-bit unsigned integer"`
	Uint32Val uint32 `env:"UINT32_VAL" flag:"uint32" default:"4294967295" desc:"32-bit unsigned integer"`
	Uint64Val uint64 `env:"UINT64_VAL" flag:"uint64" default:"18446744073709551615" desc:"64-bit unsigned integer"`
}

type DurationConfig struct {
	Timeout  time.Duration `env:"TIMEOUT" flag:"timeout,t" default:"30s" desc:"Request timeout"`
	Interval time.Duration `env:"INTERVAL" flag:"interval,i" default:"5m" desc:"Polling interval"`
	MaxWait  time.Duration `env:"MAX_WAIT" flag:"max-wait" desc:"Maximum wait time"`
}

// Helper function to reset flag.CommandLine
func resetFlagCommandLine() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}
