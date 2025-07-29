package release

import (
	_ "embed"
	"time"

	"gopkg.in/yaml.v3"
)

// Release provides self-inspection of the current build by exposing the content of release/release.yml, which should
// be updated during release.
var Release = Info{Commit: "0000000000000000000000000000000000000000"}

type Info struct {
	Tag    string
	Commit string
	Date   time.Time
}

func (i Info) String() string {
	if i.Tag != "" {
		return i.Tag
	}
	return "v0.0.0-" + i.Date.UTC().Format("20060102150405") + "-" + i.Commit[:min(12, len(i.Commit))] // go.mod pseudo-version format, though this is not required
}

func init() {
	if err := yaml.Unmarshal(releaseYaml, &Release); err != nil {
		panic(err)
	}
}

//go:embed release.yml
var releaseYaml []byte
