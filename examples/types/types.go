package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bherbruck/configlib"
)

type Config struct {
	// Integer types
	Port     int16  `env:"PORT" flag:"port,p" default:"8080" desc:"Server port"`
	MaxConns int32  `env:"MAX_CONNECTIONS" flag:"max-connections" default:"1000" desc:"Maximum connections"`
	UserID   int64  `env:"USER_ID" flag:"user-id" default:"123456789" desc:"User ID"`
	Workers  uint8  `env:"WORKERS" flag:"workers,w" default:"4" desc:"Number of workers"`
	BuffSize uint32 `env:"BUFFER_SIZE" flag:"buffer-size" default:"8192" desc:"Buffer size in bytes"`

	// Duration types
	Timeout    time.Duration `env:"TIMEOUT" flag:"timeout,t" default:"30s" desc:"Request timeout"`
	RetryDelay time.Duration `env:"RETRY_DELAY" flag:"retry-delay" default:"5s" desc:"Retry delay"`
	MaxWait    time.Duration `env:"MAX_WAIT" flag:"max-wait" default:"5m" desc:"Maximum wait time"`

	// Other types for context
	Name    string  `env:"NAME" flag:"name,n" default:"types-example" desc:"Application name"`
	Rate    float64 `env:"RATE" flag:"rate,r" default:"1.5" desc:"Processing rate"`
	Enabled bool    `env:"ENABLED" flag:"enabled,e" desc:"Enable processing"`
}

func main() {
	var cfg Config

	err := configlib.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Configuration loaded:\n")
	fmt.Printf("  Name: %s\n", cfg.Name)
	fmt.Printf("  Enabled: %t\n", cfg.Enabled)
	fmt.Printf("  Rate: %.2f\n", cfg.Rate)

	fmt.Printf("\nInteger types:\n")
	fmt.Printf("  Port: %d (int16)\n", cfg.Port)
	fmt.Printf("  Max Connections: %d (int32)\n", cfg.MaxConns)
	fmt.Printf("  User ID: %d (int64)\n", cfg.UserID)
	fmt.Printf("  Workers: %d (uint8)\n", cfg.Workers)
	fmt.Printf("  Buffer Size: %d bytes (uint32)\n", cfg.BuffSize)

	fmt.Printf("\nDuration types:\n")
	fmt.Printf("  Timeout: %v\n", cfg.Timeout)
	fmt.Printf("  Retry Delay: %v\n", cfg.RetryDelay)
	fmt.Printf("  Max Wait: %v\n", cfg.MaxWait)

	// Demonstrate usage
	fmt.Printf("\nExample usage:\n")
	if cfg.Enabled {
		fmt.Printf("  Starting server on port %d\n", cfg.Port)
		fmt.Printf("  Using %d workers with %d byte buffers\n", cfg.Workers, cfg.BuffSize)
		fmt.Printf("  Request timeout: %v\n", cfg.Timeout)
		fmt.Printf("  Will retry failed requests after %v\n", cfg.RetryDelay)
		fmt.Printf("  Maximum wait time: %v\n", cfg.MaxWait)
		fmt.Printf("  Processing rate: %.2f requests/second\n", cfg.Rate)
	} else {
		fmt.Printf("  Server is disabled\n")
	}
}
