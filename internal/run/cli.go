package run

import (
	"io"
	"os"
	"time"

	"github.com/Masterminds/semver"
	"github.com/fatih/color"
	"github.com/janmalch/roar/internal/steps"
	"github.com/janmalch/roar/models"
	"github.com/janmalch/roar/pkg/git"
	"github.com/janmalch/roar/util"
	"github.com/pkg/errors"
)

var boldGitPull = util.Bold("git pull")
var dryRunHint = color.New(color.BgWhite, color.FgBlack).Sprint(" dry-run ") + " "

func AsCli(cli models.CLI, stdout, stderr io.Writer) error {
	var err error
	var releaseAs *semver.Version

	if cli.ReleaseAs != "" {
		releaseAs, err = semver.NewVersion(cli.ReleaseAs)
		if err != nil {
			util.LogError(stderr, "Failed to parse release-as '%s' to valid semver version: %v", cli.ReleaseAs, err)
			return err
		}
	}

	dryRun := cli.DryRun

	conf, newConf, err := models.ConfigFromFile(cli.ConfigFile)
	if err != nil {
		util.LogError(stderr, "Failed to read config '%s': %v", cli.ConfigFile, err)
		return err
	}
	if newConf {
		dryRun = true
		util.LogInfo(stdout, "Created default configuration '%s' because none was found. Thus, running in dry-run mode for the first time.", cli.ConfigFile)
	}

	today := time.Now()

	r := git.NewRepo("")
	if _, err := Programmatic(r, *conf, releaseAs, today, dryRun, os.Stdout, true); err != nil {
		util.LogError(stderr, "%v", err)
		if errors.Is(err, steps.ErrRepoNotClean) {
			hasCommits, _ := r.HasCommits()
			if !hasCommits {
				util.LogInfo(stdout, "It's recommended to use %s as the message for your first commit. No conventional commit type required.", util.Bold("\"Initial commit\""))
			}
		}
		return err
	}
	return nil
}
