package configlib_test

import (
	"fmt"
	"os"

	"github.com/bherbruck/configlib"
)

func ExampleParse_multipleFlags() {
	// Set up environment and args for example
	os.Clearenv()
	os.Args = []string{"myapp", "-H", "example.com", "-p", "3000", "-d"}

	type Config struct {
		Host  string `env:"HOST" flag:"host,H" default:"localhost" desc:"Server host"`
		Port  int    `env:"PORT" flag:"port,p" default:"8080" desc:"Server port"`
		Debug bool   `env:"DEBUG" flag:"debug,d" desc:"Enable debug mode"`
	}

	var cfg Config
	err := configlib.Parse(&cfg)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Host: %s\n", cfg.Host)
	fmt.Printf("Port: %d\n", cfg.Port)
	fmt.Printf("Debug: %v\n", cfg.Debug)

	// Output:
	// Host: example.com
	// Port: 3000
	// Debug: true
}

func ExampleParse_multipleMissingFields() {
	// Clear environment to ensure no values are set
	os.Clearenv()

	// Reset CLI args
	os.Args = []string{"test"}

	// Define a config struct with multiple required fields
	type Config struct {
		APIKey      string `env:"API_KEY" flag:"api-key" required:"true"`
		DatabaseURL string `env:"DATABASE_URL" flag:"database-url" required:"true"`
		SecretKey   string `env:"SECRET_KEY" flag:"secret-key" required:"true"`
		Port        int    `env:"PORT" flag:"port" default:"8080"`
	}

	var cfg Config
	err := configlib.Parse(&cfg)

	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// missing required fields:
	//   - APIKey (env: API_KEY, flag: --api-key)
	//   - DatabaseURL (env: DATABASE_URL, flag: --database-url)
	//   - SecretKey (env: SECRET_KEY, flag: --secret-key)
}
