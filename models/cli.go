package models

import "github.com/alecthomas/kong"

type CLI struct {
	ConfigFile string           `arg:"" default:"release.toml" type:"path" help:"Path to the configuration TOML file. Default is \"./release.toml\""`
	ReleaseAs  string           `short:"x" optional:"" help:"Force a specific version, rather than computing it. It will still be verified, that this version is equal to or greater than the computed version."`
	DryRun     bool             `short:"d" help:"In dry-run mode, no changes are made to the file system or the git repository."`
	Version    kong.VersionFlag `short:"v" name:"version" help:"Print version information and quit"`
}
