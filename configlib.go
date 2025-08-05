package configlib

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type fieldInfo struct {
	EnvName     string
	CliName     string
	CliNames    []string // All CLI names (including shorthand)
	DefaultVal  string
	Required    bool
	Description string
	FieldPath   string
	Value       reflect.Value
	Type        reflect.Type
}

type Parser struct {
	fields     []fieldInfo
	flagSet    *flag.FlagSet
	flagValues map[string]string
	showHelp   bool
	boolFlags  map[string]*bool // Track boolean flags

	// Options
	disableAutoEnv  bool
	disableAutoFlag bool
	envPrefix       string
}

// Option is a functional option for configuring a Parser
type Option func(*Parser)

// NewParser creates a new parser with the given options
func NewParser(opts ...Option) *Parser {
	p := &Parser{
		flagSet:    flag.NewFlagSet("config", flag.ContinueOnError),
		fields:     make([]fieldInfo, 0),
		flagValues: make(map[string]string),
		boolFlags:  make(map[string]*bool),
	}

	// Apply options
	for _, opt := range opts {
		opt(p)
	}

	// Add help flag
	p.flagSet.BoolVar(&p.showHelp, "help", false, "Show help message")
	p.flagSet.BoolVar(&p.showHelp, "h", false, "Show help message")

	return p
}

// WithDisableAutoEnv disables automatic generation of environment variable names
func WithDisableAutoEnv() Option {
	return func(p *Parser) {
		p.disableAutoEnv = true
	}
}

// WithDisableAutoFlag disables automatic generation of CLI flag names
func WithDisableAutoFlag() Option {
	return func(p *Parser) {
		p.disableAutoFlag = true
	}
}

// WithEnvPrefix sets a prefix for all environment variable names
func WithEnvPrefix(prefix string) Option {
	return func(p *Parser) {
		p.envPrefix = prefix
	}
}

func (p *Parser) Parse(config any) error {
	// Step 1: Walk the struct and collect all fields with their metadata
	err := p.walkStruct(reflect.ValueOf(config).Elem(), "")
	if err != nil {
		return err
	}

	// Step 2: Register CLI flags based on collected fields
	p.registerFlags()

	// Step 3: Parse CLI arguments
	err = p.flagSet.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	// Check if help was requested
	if p.showHelp {
		p.PrintHelp()
		os.Exit(0)
	}

	// Process boolean flags that were set
	p.flagSet.Visit(func(f *flag.Flag) {
		if boolPtr, ok := p.boolFlags[f.Name]; ok {
			// Find the field this flag belongs to
			for _, field := range p.fields {
				for _, name := range field.CliNames {
					if name == f.Name {
						p.flagValues[field.CliName] = strconv.FormatBool(*boolPtr)
						break
					}
				}
			}
		}
	})

	// Step 4: Apply values with precedence: CLI > Env > Default
	return p.applyValues()
}

func (p *Parser) walkStruct(val reflect.Value, pathPrefix string) error {
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Build the path for nested naming
		fieldPath := fieldType.Name
		if pathPrefix != "" {
			fieldPath = pathPrefix + "." + fieldType.Name
		}

		// Handle nested structs recursively
		if field.Kind() == reflect.Struct {
			err := p.walkStruct(field, fieldPath)
			if err != nil {
				return err
			}
			continue
		}

		// Parse tags for this field
		info := p.parseFieldTags(fieldType, fieldPath, field)
		// Only add fields that have at least one way to be configured
		if info.EnvName != "" || info.CliName != "" || info.DefaultVal != "" {
			p.fields = append(p.fields, info)
		}
	}

	return nil
}

func (p *Parser) parseFieldTags(field reflect.StructField, path string, value reflect.Value) fieldInfo {
	info := fieldInfo{
		FieldPath: path,
		Value:     value,
		Type:      field.Type,
	}

	// Parse env tag
	if envTag := field.Tag.Get("env"); envTag != "" {
		info.EnvName = envTag
	} else if !p.disableAutoEnv {
		// Auto-generate from path: Server.TLS.Port -> SERVER_TLS_PORT
		info.EnvName = strings.ToUpper(strings.ReplaceAll(path, ".", "_"))
	}

	// Apply prefix to env name if set
	if info.EnvName != "" && p.envPrefix != "" {
		// Don't add prefix if the env name already starts with it
		if !strings.HasPrefix(info.EnvName, p.envPrefix) {
			info.EnvName = p.envPrefix + info.EnvName
		}
	}

	// Parse flag tag
	if flagTag := field.Tag.Get("flag"); flagTag != "" {
		// Split by comma to support multiple flags
		flags := strings.Split(flagTag, ",")
		for i, flag := range flags {
			flags[i] = strings.TrimSpace(flag)
		}
		info.CliName = flags[0] // Primary flag name
		info.CliNames = flags   // All flag names
	} else if !p.disableAutoFlag {
		// Auto-generate from path: Server.TLS.Port -> server-tls-port
		info.CliName = strings.ToLower(strings.ReplaceAll(path, ".", "-"))
		info.CliNames = []string{info.CliName}
	}

	// Parse other tags
	info.DefaultVal = field.Tag.Get("default")
	info.Required = field.Tag.Get("required") == "true"
	info.Description = field.Tag.Get("desc")

	return info
}

func (p *Parser) registerFlags() {
	for i := range p.fields {
		field := &p.fields[i] // Get pointer to avoid capturing loop variable

		// Skip if no CLI names are defined
		if len(field.CliNames) == 0 || field.CliName == "" {
			continue
		}

		// Register all flag names for this field
		for _, flagName := range field.CliNames {
			switch field.Type.Kind() {
			case reflect.String:
				p.flagSet.Func(flagName, field.Description, p.createStringHandler(field.CliName))
			case reflect.Int:
				p.flagSet.Func(flagName, field.Description, p.createIntHandler(field.CliName))
			case reflect.Float32, reflect.Float64:
				p.flagSet.Func(flagName, field.Description, p.createFloatHandler(field.CliName))
			case reflect.Bool:
				// Use BoolVar for boolean flags so they don't require a value
				boolPtr := new(bool)
				p.flagSet.BoolVar(boolPtr, flagName, false, field.Description)
				p.boolFlags[flagName] = boolPtr
			case reflect.Slice:
				p.flagSet.Func(flagName, field.Description, p.createSliceHandler(field.CliName))
			}
		}
	}

	// Set custom usage function
	p.flagSet.Usage = func() {
		p.PrintHelp()
	}
}

func (p *Parser) createStringHandler(flagName string) func(string) error {
	return func(s string) error {
		p.flagValues[flagName] = s
		return nil
	}
}

func (p *Parser) createIntHandler(flagName string) func(string) error {
	return func(s string) error {
		if _, err := strconv.Atoi(s); err != nil {
			return fmt.Errorf("invalid integer value: %s", s)
		}
		p.flagValues[flagName] = s
		return nil
	}
}

func (p *Parser) createBoolHandler(flagName string) func(string) error {
	return func(s string) error {
		// For boolean flags, if no value is provided, assume true
		if s == "" {
			p.flagValues[flagName] = "true"
			return nil
		}
		if _, err := strconv.ParseBool(s); err != nil {
			return fmt.Errorf("invalid boolean value: %s", s)
		}
		p.flagValues[flagName] = s
		return nil
	}
}

func (p *Parser) createFloatHandler(flagName string) func(string) error {
	return func(s string) error {
		if _, err := strconv.ParseFloat(s, 64); err != nil {
			return fmt.Errorf("invalid float value: %s", s)
		}
		p.flagValues[flagName] = s
		return nil
	}
}

func (p *Parser) createSliceHandler(flagName string) func(string) error {
	return func(s string) error {
		p.flagValues[flagName] = s
		return nil
	}
}

func (p *Parser) applyValues() error {
	var missingFields []string

	for _, field := range p.fields {
		var finalValue string
		var hasValue bool

		// Priority 1: CLI flags (only if non-empty and CLI name exists)
		if field.CliName != "" {
			if val, exists := p.flagValues[field.CliName]; exists && val != "" {
				finalValue = val
				hasValue = true
			}
		}

		// Priority 2: Environment variables (only if non-empty and env name exists)
		if !hasValue && field.EnvName != "" {
			if envVal := os.Getenv(field.EnvName); envVal != "" {
				finalValue = envVal
				hasValue = true
			}
		}

		// Priority 3: Default values (only if non-empty)
		if !hasValue && field.DefaultVal != "" {
			finalValue = field.DefaultVal
			hasValue = true
		}

		// Check required fields
		if field.Required && !hasValue {
			var sources []string
			if field.EnvName != "" {
				sources = append(sources, fmt.Sprintf("env: %s", field.EnvName))
			}
			if field.CliName != "" {
				sources = append(sources, fmt.Sprintf("flag: --%s", field.CliName))
			}
			missingFields = append(missingFields, fmt.Sprintf("%s (%s)",
				field.FieldPath, strings.Join(sources, ", ")))
		}

		// Set the value if we have one
		if hasValue {
			err := p.setFieldValue(field, finalValue)
			if err != nil {
				return fmt.Errorf("error setting field %s: %v", field.FieldPath, err)
			}
		}
	}

	// If there are missing required fields, return an error with all of them
	if len(missingFields) > 0 {
		return fmt.Errorf("missing required fields:\n  - %s", strings.Join(missingFields, "\n  - "))
	}

	return nil
}

func (p *Parser) setFieldValue(field fieldInfo, value string) error {
	switch field.Type.Kind() {
	case reflect.String:
		field.Value.SetString(value)
	case reflect.Int:
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		field.Value.SetInt(int64(intVal))
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.Value.SetFloat(floatVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.Value.SetBool(boolVal)
	case reflect.Slice:
		// Handle slices (e.g., comma-separated values)
		if field.Type.Elem().Kind() == reflect.String {
			parts := strings.Split(value, ",")
			slice := reflect.MakeSlice(field.Type, len(parts), len(parts))
			for i, part := range parts {
				slice.Index(i).SetString(strings.TrimSpace(part))
			}
			field.Value.Set(slice)
		}
	}
	return nil
}

// PrintHelp prints a formatted help message showing all configuration options
func (p *Parser) PrintHelp() {
	fmt.Println("Usage: " + os.Args[0] + " [options]")
	fmt.Println()
	fmt.Println("Options:")

	// Calculate max width for alignment
	maxWidth := 0
	for _, field := range p.fields {
		// Skip fields with no CLI flags
		if len(field.CliNames) == 0 || field.CliName == "" {
			continue
		}

		flagLen := 0
		for i, name := range field.CliNames {
			if i > 0 {
				flagLen += 2 // ", "
			}
			if len(name) == 1 {
				flagLen += 1 + len(name) // -x
			} else {
				flagLen += 2 + len(name) // --xxx
			}
		}
		if field.Type.Kind() != reflect.Bool {
			flagLen += 8 // " <value>"
		}
		if flagLen > maxWidth {
			maxWidth = flagLen
		}
	}
	maxWidth += 4 // padding

	// Print each field
	for _, field := range p.fields {
		// Skip fields with no CLI flags
		if len(field.CliNames) == 0 || field.CliName == "" {
			continue
		}
		p.printFieldHelp(field, maxWidth)
	}

	// Print help flag
	fmt.Printf("  -h, --help%s Show this help message\n", strings.Repeat(" ", maxWidth-10))
}

func (p *Parser) printFieldHelp(field fieldInfo, width int) {
	// Build flag string with all aliases
	var flagParts []string
	for _, name := range field.CliNames {
		if len(name) == 1 {
			flagParts = append(flagParts, "-"+name)
		} else {
			flagParts = append(flagParts, "--"+name)
		}
	}
	flag := strings.Join(flagParts, ", ")

	if field.Type.Kind() != reflect.Bool {
		flag += " <value>"
	}

	// Build description
	desc := field.Description
	if desc == "" {
		desc = field.FieldPath
	}

	// Add default value info
	if field.DefaultVal != "" && field.Type.Kind() != reflect.Bool {
		desc += fmt.Sprintf(" (default: %s)", field.DefaultVal)
	}

	// Add required marker
	if field.Required {
		desc += " [required]"
	}

	// Print formatted line
	fmt.Printf("  %-*s %s\n", width, flag, desc)
}

// GetHelp returns a help string for the configuration
func (p *Parser) GetHelp() string {
	var buf strings.Builder

	// Temporarily redirect stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	p.PrintHelp()

	w.Close()
	os.Stdout = old

	// Read the output
	output := make([]byte, 4096)
	n, _ := r.Read(output)
	buf.Write(output[:n])

	return buf.String()
}

// Parse is a convenience function to parse configuration from CLI flags, environment variables, and struct tags.
func Parse(config any) error {
	parser := NewParser()
	return parser.Parse(config)
}

// ParseWithHelp is like Parse but returns a parser instance that can be used to print help
func ParseWithHelp(config any) (*Parser, error) {
	parser := NewParser()
	err := parser.Parse(config)
	return parser, err
}
