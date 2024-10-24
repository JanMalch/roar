package run

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/fatih/color"
	"github.com/janmalch/roar/internal/steps"
	"github.com/janmalch/roar/models"
	"github.com/janmalch/roar/pkg/conventional"
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

	r := git.NewRepo("")
	if _, err := Programmatic(r, conf.File, conf.Find, conf.Replace, releaseAs, conf.Include, dryRun, os.Stdout, true); err != nil {
		util.LogError(stderr, "%v", err)
		return err
	}
	return nil
}

func Programmatic(
	r *git.Repo,
	p string,
	find string,
	replace string,
	releaseAs *semver.Version,
	includedTypes []string,
	dryrun bool,
	stdout io.Writer,
	useColor bool,
) (string, error) {
	if !useColor {
		color.NoColor = true
	}
	drp := ""
	if dryrun {
		drp = dryRunHint
	}

	path := r.PathOf(p)

	// preconditions
	if err := steps.ValidateFind(find); err != nil {
		return "", err
	}
	if err := steps.ValidateReplace(replace); err != nil {
		return "", err
	}
	if err := steps.ConfirmInputExists(path); err != nil {
		return "", err
	}
	if err := steps.ConfirmGitRepo(r); err != nil {
		return "", err
	}
	branch, err := steps.DetermineCurrentBranch(r)
	if err != nil {
		return "", err
	}
	util.LogInfo(stdout, "current branch is %s", color.New(color.Bold).Sprint(branch))
	if err := steps.ConfirmClean(r); err != nil {
		return "", err
	}

	origin, err := r.GetOrigin()
	if err != nil {
		return "", err
	}
	if origin == "" {
		util.LogInfo(stdout, "no origin is set, so %s is skipped", boldGitPull)
	} else {
		if dryrun {
			util.LogInfo(stdout, "%s%s", drp, boldGitPull)
		} else {
			if err = r.Pull(); err != nil {
				return "", errors.Wrap(err, "failed to pull to ensure freshness")
			}
			util.LogSuccess(stdout, boldGitPull)
		}
	}

	// tags
	ltag, lsemver, err := steps.DetermineLatest(r)
	if err != nil {
		return "", err
	}
	if ltag == "" {
		util.LogInfo(stdout, "no version tag found, thus assuming initial release")
	} else {
		util.LogSuccess(stdout, "determined latest version to be %s", ltag)
	}

	// handle commits
	log, err := r.CommitLogSince(ltag)
	if err != nil {
		return "", err
	}
	ccLookup, change, err := conventional.Collect(log)
	if err != nil {
		return "", err
	}
	next, err := util.Bump(lsemver, releaseAs, change)
	if err != nil {
		return "", err
	}
	ntag := "v" + next.String()
	util.LogSuccess(stdout, "determined next version to be %s", util.Bold(ntag))

	// update files
	replacement := strings.Replace(replace, "{{version}}", next.String(), 1)
	if err = steps.FindAndReplace(path, find, replacement, dryrun); err != nil {
		return "", err
	}
	util.LogSuccess(stdout, "%supdated version in %s", drp, util.Bold(p))

	if err = steps.UpdateChangelog(r.PathOf("CHANGELOG.md"), next, ccLookup, includedTypes, dryrun); err != nil {
		return "", err
	}
	util.LogSuccess(stdout, "%supdated %s", drp, util.Bold("CHANGELOG.md"))

	// commit changes
	commitMsg := fmt.Sprintf("chore(release): release version %s", ntag)
	if !dryrun {
		if err := r.Add(p, "CHANGELOG.md"); err != nil {
			return "", err
		}
		if err := r.Commit(commitMsg); err != nil {
			return "", err
		}
	}
	util.LogSuccess(stdout, "%scommited as %s", drp, util.Bold(commitMsg))
	util.LogWarning(stdout, "release commit is %s, in case you want to amend changes", util.Bold("NOT tagged yet"))

	// yay!
	util.LogSuccess(stdout, "please verify the applied changes and finalize the release by running\n\t%s", util.Bold(fmt.Sprintf("git tag %s && git push && git push --tags", ntag)))
	return ntag, nil
}
