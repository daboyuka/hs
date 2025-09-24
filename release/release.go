package release

import (
	_ "embed"
	"runtime/debug"
)

// ForkName can be overridden by forks of hs importing this package, either via init in the importing package or
// during build using:
//	-ldflags "-X github.com/daboyuka/hs/release.XXX=YYY"
var ForkName string

func ReleaseName() string {
	cmdName := "hs"
	if ForkName != "" {
		cmdName = "hs (" + ForkName + ")"
	}

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return cmdName + " v0.0.0-unknown"
	}

	return cmdName + " " + bi.Main.Version
}
