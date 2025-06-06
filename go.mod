module github.com/janmalch/roar

go 1.23.0

toolchain go1.24.1

require (
	github.com/BurntSushi/toml v1.5.0
	github.com/Masterminds/semver v1.5.0
	github.com/alecthomas/kong v1.10.0
	github.com/fatih/color v1.18.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.9.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/janmalch/roar => ./
