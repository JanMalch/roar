package main

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/janmalch/roar/internal/run"
	"github.com/janmalch/roar/models"
)

var VERSION = "0.7.0"

func main() {
	var cli models.CLI
	kong.Parse(&cli,
		kong.Name("roar"),
		kong.Description("Single-purpose CLI for opioniated semantic releases."),
		kong.Vars{
			"version": VERSION,
		},
	)

	if err := run.AsCli(cli, os.Stdout, os.Stderr); err != nil {
		os.Exit(1)
	}
}
