package config

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"

	"github.com/daboyuka/hs/hsruntime/searchpath"
	"github.com/daboyuka/hs/program/record"
	"github.com/daboyuka/hs/program/scope"
)

const (
	envPrefix = "HS_"
	filename  = ".hs"
)

func loadYAMLs(existing map[string]any) (merged map[string]any, err error) {
	return existing, searchpath.Visit(filename, func(f *os.File) error {
		return yaml.NewDecoder(f).Decode(&existing)
	})
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
	names := maps.Keys(rawVals)
	upcaseNames := make([]string, len(names))
	for i, name := range names {
		upcaseNames[i] = strings.ToUpper(name)
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
