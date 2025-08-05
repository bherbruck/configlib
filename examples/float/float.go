package main

import (
	"fmt"
	"log"

	"github.com/bherbruck/configlib"
)

type Config struct {
	// Float64 fields
	Rate      float64 `env:"RATE" flag:"rate,r" default:"1.5" desc:"Processing rate per second"`
	Threshold float64 `env:"THRESHOLD" flag:"threshold,t" default:"0.95" desc:"Success threshold (0.0-1.0)"`

	// Float32 fields
	Percentage float32 `env:"PERCENTAGE" flag:"percentage,p" desc:"Success percentage"`
	Precision  float32 `env:"PRECISION" flag:"precision" default:"0.001" desc:"Calculation precision"`

	// Other fields for context
	Name    string `env:"NAME" flag:"name,n" default:"float-example" desc:"Application name"`
	Enabled bool   `env:"ENABLED" flag:"enabled,e" desc:"Enable processing"`
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
	fmt.Printf("  Rate: %.6f (float64)\n", cfg.Rate)
	fmt.Printf("  Threshold: %.6f (float64)\n", cfg.Threshold)
	fmt.Printf("  Percentage: %.6f (float32)\n", cfg.Percentage)
	fmt.Printf("  Precision: %.6f (float32)\n", cfg.Precision)

	// Demonstrate usage
	fmt.Printf("\nExample calculations:\n")
	if cfg.Enabled {
		fmt.Printf("  Processing at %.2f items/second\n", cfg.Rate)
		fmt.Printf("  Success threshold: %.1f%%\n", cfg.Threshold*100)
		if cfg.Percentage > 0 {
			fmt.Printf("  Current success rate: %.1f%%\n", cfg.Percentage)
		}
		fmt.Printf("  Using precision: %g\n", cfg.Precision)
	} else {
		fmt.Printf("  Processing is disabled\n")
	}
}
