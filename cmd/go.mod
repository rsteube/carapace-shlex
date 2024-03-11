module github.com/carapace-sh/carapace-shlex/cmd

go 1.22.0

require (
	github.com/carapace-sh/carapace v0.50.3-0.20240311185857-480da0e3873f
	github.com/carapace-sh/carapace-bridge v0.2.16-0.20240311173237-760172108463
	github.com/carapace-sh/carapace-shlex v0.2.0
	github.com/spf13/cobra v1.8.0
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/carapace-sh/carapace-shlex => ../
