module github.com/daboyuka/hs

go 1.22

require (
	github.com/MercuryEngineering/CookieMonster v0.0.0-20180304172713-1584578b3403
	github.com/browserutils/kooky v0.2.2
	github.com/spf13/cobra v1.6.1
	github.com/spf13/pflag v1.0.5
	golang.org/x/net v0.30.0
	golang.org/x/term v0.25.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/Velocidex/json v0.0.0-20220224052537-92f3c0326e5a // indirect
	github.com/Velocidex/ordereddict v0.0.0-20230909174157-2aa49cc5d11d // indirect
	github.com/Velocidex/yaml/v2 v2.2.8 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-sqlite/sqlite3 v0.0.0-20180313105335-53dd8e640ee7 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/gonuts/binary v0.2.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/keybase/go-keychain v0.0.0-20231219164618-57a3676c3af6 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/zalando/go-keyring v0.2.5 // indirect
	golang.org/x/crypto v0.28.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	www.velocidex.com/golang/go-ese v0.2.0 // indirect
)

replace github.com/browserutils/kooky v0.2.2 => github.com/DP19/kooky v0.2.4

retract (
	v0.1.5 // force-pushed tag like a bad guy
)
