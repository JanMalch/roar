package main

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/janmalch/roar/internal/run"
	"github.com/janmalch/roar/models"
	"github.com/janmalch/roar/util"
)

var VERSION = "0.15.0"

func main() {
	var cli models.CLI
	kong.Parse(&cli,
		kong.Name("roar"),
		kong.Description("Single-purpose CLI for opioniated semantic releases.\n\n"+
			"roar is designed to be a power-user tool and expects you to be familiar with git, semantic versioning, and conventional commits. "+
			"Under the hood it's actually quite simple:\n"+
			"\n\t1. search for the latest tag starting with \"v\""+
			"\n\t2. find all commits since then"+
			"\n\t3. parse them as conventional commits"+
			"\n\t4. determine the next version"+
			"\n\t5. create a tagged commit\n"+
			"Thus, roar doesn't lock you into anything.\nFor example amending something to the release commit of a v1.0.0 is as simple as\n\t"+
			util.Bold("git tag -d v1.0.0\n\tgit add .\n\tgit commit --amend --no-edit\n\tgit tag -a \"v1.0.0\" -m \"Release v1.0.0\"")+"\n"+
			"To undo all changes, you only have to run\n\t"+util.Bold("git tag -d v1.0.0\n\tgit reset --hard HEAD^")+"\n"+
			"roar will "+util.Bold("never")+" push automatically.",
		),
		kong.Vars{
			"version": VERSION,
		},
	)

	if err := run.AsCli(cli, os.Stdout, os.Stderr); err != nil {
		os.Exit(1)
	}
}
