package main

import (
	"encoding/json"
	"log"

	"github.com/bherbruck/configlib"
)

type Config struct {
	Host     string `json:"host" env:"HOST" flag:"host" default:"localhost" desc:"Server host"`
	Port     int    `json:"port" env:"PORT" flag:"p,port" default:"8080" desc:"Server port"`
	Debug    bool   `json:"debug" env:"DEBUG" flag:"d,debug" default:"false" desc:"Enable debug mode"`
	Required string `json:"required" env:"REQUIRED" flag:"required" required:"true" desc:"Required field"`
	Server   struct {
		TLS struct {
			Enabled bool   `json:"enabled" env:"TLS_ENABLED" flag:"t,tls-enabled" default:"true" desc:"Enable TLS"`
			Cert    string `json:"cert" env:"TLS_CERT" flag:"tls-cert" required:"true" desc:"TLS certificate path"`
		} `json:"tls"`
	} `json:"server"`
}

func main() {
	// Parse config from environment variables, command line flags, and defaults
	var cfg Config
	err := configlib.Parse(&cfg)
	if err != nil {
		log.Printf("Error parsing config: %v", err)
		return
	}

	// Convert to JSON
	jsonOutput, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		panic(err)
	}

	// Print the JSON output
	println(string(jsonOutput))
}
