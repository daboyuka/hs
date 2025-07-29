package release

import (
	_ "embed"
	"time"

	"gopkg.in/yaml.v3"
)

// Release provides self-inspection of the current build by exposing the content of release/release.yml, which should
// be updated during release. This may be directly overwritten by hs forks (along with installing plugins).
var Release = Info{Commit: "0000000000000000000000000000000000000000"}

type Info struct {
	Tag    string    `yaml:"tag,omitempty"`
	Commit string    `yaml:"commit,omitempty"`
	Date   time.Time `yaml:"date,omitempty"`

	// ForkName is the name of the custom fork of hs. It's blank for this repo, but forks may override.
	ForkName string `yaml:"fork_name,omitempty"`
}

func (i Info) String() string {
	fork := "hs"
	if i.ForkName != "" {
		fork = "hs (" + i.ForkName + ")"
	}
	tag := i.Tag
	if tag == "" {
		tag = "v0.0.0-" + i.Date.UTC().Format("20060102150405") + "-" + i.Commit[:min(12, len(i.Commit))] // go.mod pseudo-version format, though this is not required
	}
	return fork + " " + tag
}

func init() {
	if err := yaml.Unmarshal(releaseYaml, &Release); err != nil {
		panic(err)
	}
}

//go:embed release.yml
var releaseYaml []byte
