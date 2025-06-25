package configlib_test

import (
	"os"
	"strings"
	"testing"

	"github.com/bherbruck/configlib"
)

func TestHelpWithMultipleFlags(t *testing.T) {
	// Clear args to avoid triggering help
	os.Args = []string{"test"}

	type Config struct {
		Host    string `env:"HOST" flag:"host,H" default:"localhost" desc:"Server host"`
		Port    int    `env:"PORT" flag:"port,p" default:"8080" desc:"Server port"`
		Debug   bool   `env:"DEBUG" flag:"debug,d" desc:"Enable debug mode"`
		Verbose bool   `env:"VERBOSE" flag:"verbose,v" desc:"Enable verbose output"`
	}

	var cfg Config
	parser, _ := configlib.ParseWithHelp(&cfg)

	helpStr := parser.GetHelp()

	// Check that help shows both short and long forms
	expectedStrings := []string{
		"--host, -H <value>",
		"--port, -p <value>",
		"--debug, -d",
		"--verbose, -v",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(helpStr, expected) {
			t.Errorf("Help output missing expected string: %s\nGot:\n%s", expected, helpStr)
		}
	}
}

func TestHelpFlag(t *testing.T) {
	// This test is skipped because we can't easily mock os.Exit
	// The help functionality is tested through other means
	t.Skip("Cannot mock os.Exit in Go")
}

func TestHelpOutput(t *testing.T) {
	type Config struct {
		Host     string `env:"HOST" flag:"host" default:"localhost" desc:"Server host"`
		Port     int    `env:"PORT" flag:"port" default:"8080" desc:"Server port"`
		Debug    bool   `env:"DEBUG" flag:"debug" default:"false" desc:"Enable debug mode"`
		Required string `env:"REQUIRED" flag:"required" required:"true" desc:"Required field"`
		Server   struct {
			TLS struct {
				Enabled bool   `env:"TLS_ENABLED" flag:"tls-enabled" default:"true" desc:"Enable TLS"`
				Cert    string `env:"TLS_CERT" flag:"tls-cert" required:"true" desc:"TLS certificate path"`
			}
		}
	}

	// Clear args to avoid triggering help
	os.Args = []string{"test"}

	var cfg Config
	parser, _ := configlib.ParseWithHelp(&cfg)

	helpStr := parser.GetHelp()

	// Check that help contains expected sections
	expectedStrings := []string{
		"Usage: test [options]",
		"Options:",
		"--host <value>",
		"Server host (default: localhost)",
		"--port <value>",
		"Server port (default: 8080)",
		"--debug",
		"Enable debug mode",
		"--required <value>",
		"Required field [required]",
		"--tls-enabled",
		"Enable TLS",
		"--tls-cert <value>",
		"TLS certificate path [required]",
		"-h, --help",
		"Show this help message",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(helpStr, expected) {
			t.Errorf("Help output missing expected string: %s", expected)
		}
	}
}

func TestParseWithHelp(t *testing.T) {
	// Clear environment
	os.Clearenv()

	// Set required field
	os.Setenv("REQUIRED", "test-value")

	// Clear args
	os.Args = []string{"test"}

	type Config struct {
		Host     string `env:"HOST" flag:"host" default:"localhost" desc:"Server hostname"`
		Required string `env:"REQUIRED" flag:"required" required:"true" desc:"API key"`
	}

	var cfg Config
	parser, err := configlib.ParseWithHelp(&cfg)

	if err != nil {
		t.Fatalf("ParseWithHelp failed: %v", err)
	}

	if parser == nil {
		t.Fatal("ParseWithHelp returned nil parser")
	}

	// Verify config was parsed correctly
	if cfg.Host != "localhost" {
		t.Errorf("Expected Host to be 'localhost', got '%s'", cfg.Host)
	}

	if cfg.Required != "test-value" {
		t.Errorf("Expected Required to be 'test-value', got '%s'", cfg.Required)
	}

	// Test that we can still call PrintHelp on the parser
	// This should not panic
	parser.PrintHelp()
}

func ExampleParser_PrintHelp() {
	type Config struct {
		Host  string `env:"HOST" flag:"host" default:"localhost" desc:"Server host"`
		Port  int    `env:"PORT" flag:"port" default:"8080" desc:"Server port"`
		Debug bool   `env:"DEBUG" flag:"debug" desc:"Enable debug mode"`
	}

	// Clear args to avoid triggering help
	os.Args = []string{"myapp"}

	var cfg Config
	parser, _ := configlib.ParseWithHelp(&cfg)
	parser.PrintHelp()

	// Output:
	// Usage: myapp [options]
	//
	// Options:
	//   --host <value>     Server host (default: localhost)
	//   --port <value>     Server port (default: 8080)
	//   --debug            Enable debug mode
	//   -h, --help         Show this help message
}
