# configlib

A Go library for parsing configuration from environment variables, CLI flags, and struct tags with support for nested structs and comprehensive error reporting.

## Features

- **Multiple sources**: Parse configuration from environment variables, CLI flags, and default values
- **Priority order**: CLI flags > Environment variables > Default values
- **Nested struct support**: Automatically handle nested configuration structures
- **Type safety**: Support for string, int, bool, and string slice types
- **Required fields**: Mark fields as required and get comprehensive error messages
- **Auto-naming**: Automatic generation of env var and CLI flag names from struct field paths
- **Configurable auto-naming**: Optionally disable auto-generation of env vars or flags
- **Environment variable prefixes**: Add custom prefixes to all environment variables
- **Comprehensive error reporting**: Collects all missing required fields and reports them together
- **Built-in help**: Automatic help generation with `--help` or `-h` flags

## Installation

```bash
go get github.com/bherbruck/configlib
```

## Usage

### Basic Example

```go
package main

import (
    "fmt"
    "log"
    "github.com/bherbruck/configlib"
)

type Config struct {
    Host     string `env:"HOST" cli:"host" default:"localhost" desc:"Server host"`
    Port     int    `env:"PORT" cli:"port" default:"8080" desc:"Server port"`
    Debug    bool   `env:"DEBUG" cli:"debug" default:"false" desc:"Enable debug mode"`
    APIKey   string `env:"API_KEY" cli:"api-key" required:"true" desc:"API key"`
}

func main() {
    var cfg Config
    if err := configlib.Parse(&cfg); err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Config: %+v\n", cfg)
}
```

### Multiple Flags (Shorthand)

You can specify multiple flags for a single field using comma-separated values in the `cli` tag:

```go
type Config struct {
    Host    string `env:"HOST" cli:"host,H" desc:"Server host"`
    Port    int    `env:"PORT" cli:"port,p" desc:"Server port"`
    Debug   bool   `env:"DEBUG" cli:"debug,d" desc:"Enable debug mode"`
    Verbose bool   `env:"VERBOSE" cli:"verbose,v" desc:"Enable verbose output"`
}
```

This allows both long and short forms:
```bash
# Long form
./myapp --host example.com --port 8080 --debug

# Short form
./myapp -H example.com -p 8080 -d

# Mixed
./myapp --host example.com -p 8080 -d
```

### Nested Structs

```go
type Config struct {
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
```

### Comprehensive Error Reporting

When multiple required fields are missing, configlib collects all errors and reports them together:

```
missing required fields:
  - Server.Host (env: SERVER_HOST, cli: --server-host)
  - Database.Host (env: DB_HOST, cli: --db-host)
  - Database.Password (env: DB_PASSWORD, cli: --db-password)
```

This makes it easy to identify all missing configuration at once, rather than discovering them one at a time.

## Struct Tags

- `env`: Name of the environment variable (auto-generated if not specified)
- `cli`: Name of the CLI flag (auto-generated if not specified). Supports multiple flags separated by commas (e.g., `cli:"host,h"` for both `--host` and `-h`)
- `default`: Default value if not provided via env or CLI
- `required`: Set to "true" to make the field required
- `desc`: Description for the CLI flag help text

## Auto-naming Convention

If you don't specify `env` or `cli` tags, they are automatically generated:

- Environment variables: `FieldPath` → `FIELD_PATH` (e.g., `Server.Host` → `SERVER_HOST`)
- CLI flags: `FieldPath` → `field-path` (e.g., `Server.Host` → `server-host`)

### Disabling Auto-generation

You can disable automatic generation of environment variables or CLI flags:

```go
// Disable auto-generation of environment variables
parser := configlib.NewParser().WithDisableAutoEnv()
err := parser.Parse(&cfg)

// Disable auto-generation of CLI flags
parser := configlib.NewParser().WithDisableAutoFlag()
err := parser.Parse(&cfg)

// Disable both
parser := configlib.NewParser().
    WithDisableAutoEnv().
    WithDisableAutoFlag()
err := parser.Parse(&cfg)
```

When auto-generation is disabled:
- Fields without explicit `env` tags won't be configurable via environment variables
- Fields without explicit `cli` tags won't be configurable via CLI flags
- Fields can still use default values

### Environment Variable Prefixes

You can add a prefix to all environment variable names:

```go
// Add "MYAPP_" prefix to all env vars
parser := configlib.NewParser().WithEnvPrefix("MYAPP_")
err := parser.Parse(&cfg)
```

This affects both auto-generated and explicitly defined environment variable names:
- `HOST` → `MYAPP_HOST`
- `Server.Port` → `MYAPP_SERVER_PORT`
- Already prefixed names are not double-prefixed

Example:
```go
type Config struct {
    Host string `env:"HOST"`        // Will become MYAPP_HOST
    Port int                        // Will become MYAPP_PORT (auto-generated)
    Database struct {
        Name string                 // Will become MYAPP_DATABASE_NAME
    }
}

parser := configlib.NewParser().WithEnvPrefix("MYAPP_")
```

## Supported Types

- `string`
- `int`
- `bool`
- `[]string` (comma-separated values)

## Priority Order

Values are resolved in the following order (highest priority first):

1. CLI flags
2. Environment variables
3. Default values

## Help Functionality

The library automatically provides help functionality through the `--help` or `-h` flags:

```bash
./myapp --help
```

This will display:
- All configuration options organized by groups
- Environment variable names
- CLI flag names
- Types, default values, and descriptions
- Required field indicators

### Programmatic Help Access

You can also access help programmatically:

```go
parser, err := configlib.ParseWithHelp(&cfg)
if err != nil {
    // Handle error
}

// Print help manually
parser.PrintHelp()

// Or get help as a string
helpText := parser.GetHelp()
```

## Advanced Configuration Options

### Parser Options

The library provides several options to customize parsing behavior:

```go
// Create a parser with custom options
parser := configlib.NewParser().
    WithDisableAutoEnv().      // Disable auto-generation of env var names
    WithDisableAutoFlag().     // Disable auto-generation of CLI flag names
    WithEnvPrefix("MYAPP_")    // Add prefix to all env var names

err := parser.Parse(&cfg)
```

### Complete Example with Options

```go
package main

import (
    "fmt"
    "log"
    "github.com/bherbruck/configlib"
)

type Config struct {
    // Only configurable via CLI flag (auto-env disabled)
    Host string `cli:"host" default:"localhost"`
    
    // Only configurable via env var (auto-flag disabled)
    Port int `env:"PORT" default:"8080"`
    
    // Explicitly defined both
    Debug bool `env:"DEBUG" cli:"debug,d"`
    
    // Will have MYAPP_ prefix: MYAPP_DATABASE_URL
    Database struct {
        URL string `env:"DATABASE_URL"`
    }
}

func main() {
    var cfg Config
    
    parser := configlib.NewParser().
        WithDisableAutoEnv().
        WithDisableAutoFlag().
        WithEnvPrefix("MYAPP_")
    
    if err := parser.Parse(&cfg); err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Config: %+v\n", cfg)
}
```

## License

MIT
