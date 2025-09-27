package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/daboyuka/hs/hsruntime/cookie"
	"github.com/daboyuka/hs/hsruntime/searchpath"
	"github.com/daboyuka/hs/program/record"
	"github.com/daboyuka/hs/program/scope"
)

const (
	envPrefix = "HS_"
	filename  = ".hs"
)

var defaultBaseConfiguration = `# Load cookies from the following browsers to use browser-based authentication
browser_loaders:`

func init() {
	for _, browser := range cookie.AllSupportedBrowsers() {
		switch browser {
		case "chrome", "firefox":
			defaultBaseConfiguration += fmt.Sprintf("\n    - %s", browser)
		default:
			defaultBaseConfiguration += fmt.Sprintf("\n#    - %s", browser)
		}
	}
}

func loadYAMLs(existing map[string]any) (merged map[string]any, err error) {
	return existing, searchpath.Visit(filename, func(f *os.File) error {
		return yaml.NewDecoder(f).Decode(&existing)
	})
}

func WarnMissingBaseConfiguration() error {
	if home, err := os.UserHomeDir(); err == nil {
		configPath := filepath.Join(home, filename)
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return fmt.Errorf("config file %s does not exist; run `hs init` to create it", configPath)
		}
	}
	return nil
}

func DefaultConfiguration(_ string) (string, error) { return defaultBaseConfiguration, nil }

func CreateConfigurationFile(cfg string) (string, error) {
	// Only if it doesn't exist yet, and only in the user's home directory; do not search path
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine user's home directory to create config file")
	}

	configPath := filepath.Join(home, filename)
	_, err = os.Stat(configPath)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("config file %s already exists", configPath)
	}

	f, err := os.OpenFile(configPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644)
	if err != nil {
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			// still does not exist; some error other than losing the race
			return "", fmt.Errorf("failed to create config file %s: %w", configPath, err)
		}
		return "", fmt.Errorf("another process created the config file %s while we were trying to create it: %w", configPath, err)
	}
	defer f.Close()

	var document yaml.Node
	if err := yaml.Unmarshal([]byte(cfg), &document); err != nil {
		panic(fmt.Errorf("failed to parse default config: %w", err))
	} else if _, err := f.Write([]byte(cfg)); err != nil {
		return "", fmt.Errorf("failed to write initial config: %w", err)
	}
	return fmt.Sprintf("config file %s has been created with contents:\n\n%s\n", configPath, cfg), nil
}

func loadEnv(existing map[string]any) (merged map[string]any, err error) {
	merged = existing
	if merged == nil {
		merged = make(map[string]any)
	}
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, envPrefix) {
			continue
		}
		idx := strings.IndexRune(env, '=')
		if idx == -1 {
			panic(fmt.Errorf("missing equals in environment variable somehow: %s", env))
		}

		key, value := env[len(envPrefix):idx], env[idx+1:]
		merged[key] = value
	}
	return merged, nil
}

// Load loads all configuration into a new child scope/bindings derived from the given scope/bindings.
func Load(scp *scope.Scope, binds *scope.Bindings) (*scope.Scope, *scope.Bindings, error) {
	rawVals, err := loadYAMLs(nil)
	if err != nil {
		return nil, nil, err
	}
	rawVals, err = loadEnv(rawVals) // override YAMLs with env vars
	if err != nil {
		return nil, nil, err
	}

	// Collect and uppercase all names
	names := make([]string, 0, len(rawVals))
	upcaseNames := make([]string, 0, len(rawVals))
	for name := range rawVals {
		names = append(names, name)
		upcaseNames = append(upcaseNames, strings.ToUpper(name))
	}

	// Convert (uppercase) names to unique Idents
	nextScp, ids := scope.NewScope(scp, upcaseNames...)

	// Join new Idents with config values
	vals := make(map[scope.Ident]record.Record)
	for i, name := range names {
		vals[ids[i]] = rawVals[name]
	}

	return nextScp, scope.NewBindings(binds, vals), nil
}
