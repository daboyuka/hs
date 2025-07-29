package release

import (
	_ "embed"
	"time"
)

// These variables are the source for this build's release info. They should be overridden during build using:
//	-ldflags "-X github.com/daboyuka/hs/release.XXX=YYY"
var (
	Tag      string
	Commit   string
	DateRaw  string
	ForkName string

	// Date is DateRaw parsed as RFC3339
	Date time.Time
)

func init() {
	Date, _ = time.Parse(time.RFC3339, DateRaw)
}

func ReleaseName() string {
	cmdName := "hs"
	if ForkName != "" {
		cmdName = "hs (" + ForkName + ")"
	}
	tag := Tag
	if tag == "" {
		commit := Commit
		if commit == "" {
			commit = "0000000000000000000000000000000000000000"
		}
		tag = "v0.0.0-" + Date.UTC().Format("20060102150405") + "-" + commit[:min(12, len(commit))] // go.mod pseudo-version format, though this is not required
	}
	return cmdName + " " + tag
}
