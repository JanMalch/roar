package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/janmalch/roar/internal/run"
	"github.com/janmalch/roar/internal/steps"
	"github.com/janmalch/roar/models"
	"github.com/janmalch/roar/pkg/git"
	"github.com/janmalch/roar/util"
)

var VERSION = "0.17.0"

func main() {
	// Using current repo version for help text, if people copy the commands
	r := git.NewRepo("")
	tag, _, _ := steps.DetermineLatest(r)
	if tag == "" {
		tag = "v1.0.0"
	}

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
			fmt.Sprintf("Thus, roar doesn't lock you into anything.\nFor example amending something to the release commit of a %s is as simple as\n\t", tag)+
			util.Bold(fmt.Sprintf("git tag -d %s\n\tgit add .\n\tgit commit --amend --no-edit\n\tgit tag -a \"%s\" -m \"Release %s\"", tag, tag, tag))+"\n"+
			"To undo all changes, you only have to run\n\t"+util.Bold(fmt.Sprintf("git tag -d %s\n\tgit reset --hard HEAD^", tag))+"\n"+
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
