package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v77/github"
	"github.com/janmalch/roar/internal/run"
	"github.com/janmalch/roar/internal/steps"
	"github.com/janmalch/roar/models"
	"github.com/janmalch/roar/pkg/git"
	"github.com/janmalch/roar/util"
)

var VERSION = "0.18.1"

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

	err := run.AsCli(cli, os.Stdout, os.Stderr)
	checkForUpdate()
	if err != nil {
		os.Exit(1)
	}
}

func checkForUpdate() {
	client := github.NewClient(nil)
	ctx := context.Background()
	release, _, err := client.Repositories.GetLatestRelease(ctx, "JanMalch", "roar")
	if err != nil {
		util.LogWarning(os.Stdout, "Failed to check for updates.")
		return
	}
	latestTag := "v" + VERSION
	if latestTag == release.GetTagName() {
		return
	}

	url := release.GetHTMLURL()
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Align(lipgloss.Left)

	fmt.Println(style.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Inline(true).Bold(true).Render("A new roar version is available!\n"),
			fmt.Sprintf("%s -> %s", latestTag, release.GetTagName()),
			lipgloss.NewStyle().Inline(true).Foreground(lipgloss.Color("12")).Render(url),
		),
	))

}
