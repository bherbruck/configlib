package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/bherbruck/configlib"
)

// Example 1: Disable auto-generation of env vars
type ConfigNoAutoEnv struct {
	// This field will NOT have an auto-generated env var name
	Host string `json:"host" flag:"host" default:"localhost" desc:"Server host"`

	// This field WILL have the explicit env var name
	Port int `json:"port" env:"PORT" flag:"port" default:"8080" desc:"Server port"`

	// This field will have no env var at all
	Debug bool `json:"debug" flag:"debug" default:"false" desc:"Enable debug mode"`
}

// Example 2: Disable auto-generation of CLI flags
type ConfigNoAutoFlag struct {
	// This field will NOT have an auto-generated CLI flag
	Host string `json:"host" env:"HOST" default:"localhost" desc:"Server host"`

	// This field WILL have the explicit CLI flag
	Port int `json:"port" env:"PORT" flag:"port" default:"8080" desc:"Server port"`

	// This field will have no CLI flag at all
	Debug bool `json:"debug" env:"DEBUG" default:"false" desc:"Enable debug mode"`
}

// Example 3: Add prefix to env vars
type ConfigWithPrefix struct {
	Host  string `json:"host" env:"HOST" flag:"host" default:"localhost" desc:"Server host"`
	Port  int    `json:"port" env:"PORT" flag:"port" default:"8080" desc:"Server port"`
	Debug bool   `json:"debug" flag:"debug" default:"false" desc:"Enable debug mode"`

	// Nested struct - auto-generated env names will also get the prefix
	Database struct {
		Name string `json:"name" flag:"db-name" desc:"Database name"`
		User string `json:"user" flag:"db-user" desc:"Database user"`
	} `json:"database"`
}

func main() {
	fmt.Println("=== Example 1: Disable Auto-Generation of Environment Variables ===")
	var cfg1 ConfigNoAutoEnv
	parser1 := configlib.NewParser(configlib.WithDisableAutoEnv())
	err := parser1.Parse(&cfg1)
	if err != nil {
		log.Printf("Error parsing config: %v", err)
	}

	jsonOutput, _ := json.MarshalIndent(cfg1, "", "  ")
	fmt.Println("Config:", string(jsonOutput))
	fmt.Println()

	fmt.Println("=== Example 2: Disable Auto-Generation of CLI Flags ===")
	var cfg2 ConfigNoAutoFlag
	parser2 := configlib.NewParser(configlib.WithDisableAutoFlag())
	err = parser2.Parse(&cfg2)
	if err != nil {
		log.Printf("Error parsing config: %v", err)
	}

	jsonOutput, _ = json.MarshalIndent(cfg2, "", "  ")
	fmt.Println("Config:", string(jsonOutput))
	fmt.Println()

	fmt.Println("=== Example 3: Add Prefix to Environment Variables ===")
	fmt.Println("Environment variables will be prefixed with 'MYAPP_'")
	fmt.Println("e.g., HOST -> MYAPP_HOST, Database.Name -> MYAPP_DATABASE_NAME")
	var cfg3 ConfigWithPrefix
	parser3 := configlib.NewParser(configlib.WithEnvPrefix("MYAPP_"))
	err = parser3.Parse(&cfg3)
	if err != nil {
		log.Printf("Error parsing config: %v", err)
	}

	jsonOutput, _ = json.MarshalIndent(cfg3, "", "  ")
	fmt.Println("Config:", string(jsonOutput))
	fmt.Println()

	fmt.Println("=== Example 4: Combine Multiple Options ===")
	fmt.Println("Disable auto-flags AND add env prefix")
	var cfg4 ConfigWithPrefix
	parser4 := configlib.NewParser(
		configlib.WithDisableAutoFlag(),
		configlib.WithEnvPrefix("APP_"),
	)
	err = parser4.Parse(&cfg4)
	if err != nil {
		log.Printf("Error parsing config: %v", err)
	}

	jsonOutput, _ = json.MarshalIndent(cfg4, "", "  ")
	fmt.Println("Config:", string(jsonOutput))
}
