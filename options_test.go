package configlib_test

import (
	"os"
	"testing"

	"github.com/bherbruck/configlib"
)

func TestDisableAutoEnv(t *testing.T) {
	type Config struct {
		// Should NOT have auto-generated env var
		Field1 string `flag:"field1" default:"default1"`
		// Should have explicit env var
		Field2 string `env:"EXPLICIT_ENV" flag:"field2"`
		// Nested struct
		Nested struct {
			// Should NOT have auto-generated env var
			Field3 string `flag:"field3"`
		}
	}

	// Set environment variables
	os.Setenv("FIELD1", "from_env1")
	os.Setenv("EXPLICIT_ENV", "from_env2")
	os.Setenv("NESTED_FIELD3", "from_env3")
	defer func() {
		os.Unsetenv("FIELD1")
		os.Unsetenv("EXPLICIT_ENV")
		os.Unsetenv("NESTED_FIELD3")
	}()

	// Save and restore os.Args
	oldArgs := os.Args
	os.Args = []string{"test"}
	defer func() { os.Args = oldArgs }()

	var cfg Config
	parser := configlib.NewParser(configlib.WithDisableAutoEnv())
	err := parser.Parse(&cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Field1 should use default (env var ignored due to disabled auto-env)
	if cfg.Field1 != "default1" {
		t.Errorf("Field1: expected 'default1', got '%s'", cfg.Field1)
	}

	// Field2 should use env var (explicit env tag)
	if cfg.Field2 != "from_env2" {
		t.Errorf("Field2: expected 'from_env2', got '%s'", cfg.Field2)
	}

	// Nested.Field3 should be empty (env var ignored)
	if cfg.Nested.Field3 != "" {
		t.Errorf("Nested.Field3: expected empty, got '%s'", cfg.Nested.Field3)
	}
}

func TestDisableAutoFlag(t *testing.T) {
	type Config struct {
		// Should NOT have auto-generated CLI flag
		Field1 string `env:"FIELD1" default:"default1"`
		// Should have explicit CLI flag
		Field2 string `env:"FIELD2" flag:"field2"`
	}

	// Test with CLI args - only field2 should work
	oldArgs := os.Args
	os.Args = []string{"test", "--field2", "from_cli2"}
	defer func() { os.Args = oldArgs }()

	var cfg Config
	parser := configlib.NewParser(configlib.WithDisableAutoFlag())
	err := parser.Parse(&cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Field1 should use default (no CLI flag due to disabled auto-flag)
	if cfg.Field1 != "default1" {
		t.Errorf("Field1: expected 'default1', got '%s'", cfg.Field1)
	}

	// Field2 should use CLI value (explicit flag tag)
	if cfg.Field2 != "from_cli2" {
		t.Errorf("Field2: expected 'from_cli2', got '%s'", cfg.Field2)
	}
}

func TestEnvPrefix(t *testing.T) {
	type Config struct {
		// Explicit env name
		Field1 string `env:"HOST"`
		// Auto-generated env name
		Field2 string
		// Nested struct
		Database struct {
			Name string `env:"DB_NAME"`
			Port int
		}
	}

	// Set environment variables with prefix
	os.Setenv("MYAPP_HOST", "myhost")
	os.Setenv("MYAPP_FIELD2", "myfield2")
	os.Setenv("MYAPP_DB_NAME", "mydb")
	os.Setenv("MYAPP_DATABASE_PORT", "5432")
	defer func() {
		os.Unsetenv("MYAPP_HOST")
		os.Unsetenv("MYAPP_FIELD2")
		os.Unsetenv("MYAPP_DB_NAME")
		os.Unsetenv("MYAPP_DATABASE_PORT")
	}()

	// Save and restore os.Args
	oldArgs := os.Args
	os.Args = []string{"test"}
	defer func() { os.Args = oldArgs }()

	var cfg Config
	parser := configlib.NewParser(configlib.WithEnvPrefix("MYAPP_"))
	err := parser.Parse(&cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if cfg.Field1 != "myhost" {
		t.Errorf("Field1: expected 'myhost', got '%s'", cfg.Field1)
	}

	if cfg.Field2 != "myfield2" {
		t.Errorf("Field2: expected 'myfield2', got '%s'", cfg.Field2)
	}

	if cfg.Database.Name != "mydb" {
		t.Errorf("Database.Name: expected 'mydb', got '%s'", cfg.Database.Name)
	}

	if cfg.Database.Port != 5432 {
		t.Errorf("Database.Port: expected 5432, got %d", cfg.Database.Port)
	}
}

func TestEnvPrefixNoDoublePrefixing(t *testing.T) {
	type Config struct {
		// Already has the prefix
		Field1 string `env:"MYAPP_HOST"`
		// Doesn't have the prefix
		Field2 string `env:"PORT"`
	}

	// Set environment variables
	os.Setenv("MYAPP_HOST", "host1")
	os.Setenv("MYAPP_PORT", "8080")
	defer func() {
		os.Unsetenv("MYAPP_HOST")
		os.Unsetenv("MYAPP_PORT")
	}()

	// Save and restore os.Args
	oldArgs := os.Args
	os.Args = []string{"test"}
	defer func() { os.Args = oldArgs }()

	var cfg Config
	parser := configlib.NewParser(configlib.WithEnvPrefix("MYAPP_"))
	err := parser.Parse(&cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should not double-prefix
	if cfg.Field1 != "host1" {
		t.Errorf("Field1: expected 'host1', got '%s'", cfg.Field1)
	}

	// Should add prefix
	if cfg.Field2 != "8080" {
		t.Errorf("Field2: expected '8080', got '%s'", cfg.Field2)
	}
}

func TestCombinedOptions(t *testing.T) {
	type Config struct {
		// Only env var (auto-flag disabled)
		Field1 string `env:"FIELD1"`
		// Only CLI flag (auto-env disabled)
		Field2 string `flag:"field2"`
		// Both explicit
		Field3 string `env:"FIELD3" flag:"field3"`
		// Default only (both auto disabled)
		Field4 string `default:"default4"`
	}

	// Set environment variables with prefix
	os.Setenv("APP_FIELD1", "env1")
	os.Setenv("APP_FIELD3", "env3")
	defer func() {
		os.Unsetenv("APP_FIELD1")
		os.Unsetenv("APP_FIELD3")
	}()

	// Set CLI args
	oldArgs := os.Args
	os.Args = []string{"test", "--field2", "cli2", "--field3", "cli3"}
	defer func() { os.Args = oldArgs }()

	var cfg Config
	parser := configlib.NewParser(
		configlib.WithDisableAutoEnv(),
		configlib.WithDisableAutoFlag(),
		configlib.WithEnvPrefix("APP_"),
	)
	err := parser.Parse(&cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if cfg.Field1 != "env1" {
		t.Errorf("Field1: expected 'env1', got '%s'", cfg.Field1)
	}

	if cfg.Field2 != "cli2" {
		t.Errorf("Field2: expected 'cli2', got '%s'", cfg.Field2)
	}

	// CLI should override env
	if cfg.Field3 != "cli3" {
		t.Errorf("Field3: expected 'cli3', got '%s'", cfg.Field3)
	}

	if cfg.Field4 != "default4" {
		t.Errorf("Field4: expected 'default4', got '%s'", cfg.Field4)
	}
}
