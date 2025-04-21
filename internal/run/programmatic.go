package run

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/fatih/color"
	"github.com/janmalch/roar/internal/steps"
	"github.com/janmalch/roar/models"
	"github.com/janmalch/roar/pkg/conventional"
	"github.com/janmalch/roar/pkg/git"
	"github.com/janmalch/roar/util"
	"github.com/pkg/errors"
)

func Programmatic(
	r *git.Repo,
	c models.Config,
	releaseAs *semver.Version,
	today time.Time,
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

	// preconditions
	for _, u := range c.Updates {
		path := r.PathOf(u.File)
		if err := steps.ValidateFind(u.Find); err != nil {
			return "", err
		}
		if err := steps.ValidateReplace(u.Replace); err != nil {
			return "", err
		}
		if err := steps.ConfirmInputExists(path); err != nil {
			return "", err
		}
	}
	if err := steps.ConfirmGitRepo(r); err != nil {
		return "", err
	}
	if err := steps.ConfirmClean(r); err != nil {
		return "", err
	}
	branch, err := steps.ValidateBranch(r, c.Branch)
	if err != nil {
		return "", err
	}
	if c.Branch != "" {
		util.LogSuccess(stdout, "current branch %s matches %s pattern", color.New(color.Bold).Sprint(branch), color.New(color.Bold).Sprint(c.Branch))
	}

	// tags
	ltag, lsemver, err := steps.DetermineLatest(r)
	if err != nil {
		return "", err
	}
	if ltag == "" {
		util.LogInfo(stdout, "no version tag found, thus assuming initial release")
	} else {
		util.LogInfo(stdout, "determined latest version to be %s", ltag)
	}
	util.LogEmptyLine(stdout)

	// handle commits
	log, err := r.CommitLogSince(ltag)
	if err != nil {
		return "", err
	}
	ccLookup, change, err := conventional.Collect(log)
	if err != nil {
		if !errors.Is(err, conventional.ErrNoCommits) || releaseAs == nil {
			return "", err
		}
	}
	next, err := util.Bump(lsemver, releaseAs, change)
	if err != nil {
		return "", err
	}
	ntag := "v" + next.String()
	util.LogSuccess(stdout, "determined next version to be %s", util.Bold(ntag))

	// update files
	if len(c.Updates) == 0 {
		util.LogWarning(stdout, "No update instructions defined in configuration file. Did you forget to add at least one [[update]] section?")
	} else {
		epoch := fmt.Sprintf("%d", time.Now().UnixMilli())
		for _, u := range c.Updates {
			path := r.PathOf(u.File)
			// create replacement string: u.Replace is the template from the config
			replacement := strings.ReplaceAll(strings.ReplaceAll(u.Replace, "{{version}}", next.String()), "{{epoch}}", epoch)
			// update file with replacement string
			if err = steps.FindAndReplace(path, u.Find, replacement, dryrun); err != nil {
				return "", err
			}
			util.LogSuccess(stdout, "%supdated %s", drp, util.Bold(u.File))
		}
	}

	if err = steps.UpdateChangelog(r.PathOf("CHANGELOG.md"), &c.Changelog, next, lsemver, ccLookup, today, dryrun); err != nil {
		return "", err
	}
	util.LogSuccess(stdout, "%supdated %s", drp, util.Bold("CHANGELOG.md"))

	// commit changes
	commitMsg := fmt.Sprintf("chore(release): release version %s", ntag)
	if !dryrun {
		for _, u := range c.Updates {
			if err := r.Add(u.File); err != nil {
				return "", err
			}
		}
		if err := r.Add("CHANGELOG.md"); err != nil {
			return "", err
		}
		if err := r.Commit(commitMsg); err != nil {
			return "", err
		}
	}
	util.LogSuccess(stdout, "%scommited as %s", drp, util.Bold(commitMsg))

	// TODO: make this step configurable, but enabled by default
	if !dryrun {
		if err = r.AddTag(ntag); err != nil {
			return "", err
		}
	}
	util.LogSuccess(stdout, "%stagged as %s", drp, util.Bold(ntag))
	util.LogEmptyLine(stdout)

	// yay!
	util.LogInfo(stdout, "please verify the applied changes and finalize the release by running\n\t%s", util.Bold("git push && git push --tags"))
	return ntag, nil
}
