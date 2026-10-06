module github.com/mvelosop/vloop

go 1.27.1

require (
	github.com/BurntSushi/toml v1.6.0
	github.com/bmatcuk/doublestar/v4 v4.10.2
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.3
	github.com/spf13/cobra v1.10.2
	github.com/spf13/pflag v1.0.9
	golang.org/x/sys v0.46.0
	golang.org/x/text v0.14.0
)

require github.com/inconshreveable/mousetrap v1.1.0 // indirect

// v1.0.0 and the v2.0.0 tags were the numbering before the repository went
// public, and were removed; this line continues as v0.8.0.
retract v1.0.0
