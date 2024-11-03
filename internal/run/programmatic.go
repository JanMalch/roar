package run

import (
	"fmt"
	"io"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/fatih/color"
	"github.com/janmalch/roar/internal/steps"
	"github.com/janmalch/roar/pkg/conventional"
	"github.com/janmalch/roar/pkg/git"
	"github.com/janmalch/roar/util"
	"github.com/pkg/errors"
)

func Programmatic(
	r *git.Repo,
	p string,
	find string,
	replace string,
	releaseAs *semver.Version,
	includedTypes []string,
	gitService string,
	gitUrl string,
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
	if err := steps.ConfirmClean(r); err != nil {
		return "", err
	}
	branch, err := steps.DetermineCurrentBranch(r)
	if err != nil {
		return "", err
	}
	util.LogInfo(stdout, "current branch is %s", color.New(color.Bold).Sprint(branch))

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

	gs, known := steps.DetectGitService(origin, gitUrl, gitService)
	if !known {
		util.LogInfo(stdout, "no git service detected; configure \"gitService\" and possibly \"gitServiceUrl\" to enable links in the changelog")
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
		// TODO: support releases without commits somehow? If user just wants to bump the version to e.g. v1.0.0 after v0 phase
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

	if err = steps.UpdateChangelog(r.PathOf("CHANGELOG.md"), gs, next, lsemver, ccLookup, includedTypes, dryrun); err != nil {
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
