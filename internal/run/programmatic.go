package run

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
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

var ErrPatternNoMatches = errors.New("pattern matches zero files")
var ErrHooksNotAllowed = errors.New("there are hooks configured, but roar was started without the allow-hooks flag.")

func Programmatic(
	r *git.Repo,
	c models.Config,
	releaseAs *semver.Version,
	today time.Time,
	dryrun bool,
	allowDirty bool,
	allowHooks bool,
	stdout io.Writer,
	useColor bool,
) (string, error) {
	if allowDirty && !dryrun {
		return "", ErrBadAllowDirty
	}
	if !useColor {
		color.NoColor = true
	}
	if c.Hooks != nil && c.Hooks.BeforeStaging.Cmd != "" && !allowHooks {
		return "", ErrHooksNotAllowed
	}
	drp := ""
	if dryrun {
		drp = dryRunHint
	}

	// preconditions
	for _, u := range c.Updates {
		if err := steps.ValidateFind(u.Find); err != nil {
			return "", err
		}
		if err := steps.ValidateReplace(u.Replace); err != nil {
			return "", err
		}
		if u.File != "" {
			path := r.PathOf(u.File)
			if err := steps.ConfirmInputExists(path); err != nil {
				return "", err
			}
		}
	}
	if err := steps.ConfirmGitRepo(r); err != nil {
		return "", err
	}
	if err := steps.ConfirmClean(r); err != nil {
		if allowDirty {
			util.LogError(stdout, "%s%s", drp, err)
		} else {
			return "", err
		}
	}
	branch, err := steps.ValidateBranch(r, c.Branch)
	if err != nil {
		return "", err
	}
	if c.Branch != "" {
		util.LogSuccess(stdout, "current branch %s matches %s pattern", color.New(color.Bold).Sprint(branch), color.New(color.Bold).Sprint(c.Branch))
	}
	if !dryrun {
		err = r.FetchTags()
		if err != nil {
			return "", err
		}
	}
	util.LogSuccess(stdout, "%s%s", drp, util.Bold("get fetch --tags"))

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
			if u.File != "" {
				err = updateFile(r.PathOf(u.File), u.Find, u.Replace, next, epoch, drp, dryrun, stdout)
				if err != nil {
					return "", err
				}
			}
			if u.Pattern != "" {
				matches, err := filepath.Glob(u.Pattern)
				if err != nil {
					return "", err
				}
				if len(matches) == 0 {
					return "", ErrPatternNoMatches
				}
				for _, match := range matches {
					err = updateFile(match, u.Find, u.Replace, next, epoch, drp, dryrun, stdout)
					if err != nil {
						return "", err
					}
				}
			}
		}
	}
	if c.Npm != nil {
		cmd, err := npmVersion(r.Dir, next, *c.Npm, dryrun)
		if err != nil {
			return "", err
		}
		util.LogSuccess(stdout, "%s%s", drp, util.Bold(cmd))
		if !dryrun {
			modified, err := r.ListModified()
			if err != nil {
				return "", err
			}
			for _, file := range modified {
				if strings.HasSuffix(file, "package.json") {
					err = r.Add(file)
					if err != nil {
						return "", err
					}
				}
			}
		}
	}

	if err = steps.UpdateChangelog(r.PathOf("CHANGELOG.md"), &c.Changelog, next, lsemver, ccLookup, today, dryrun); err != nil {
		return "", err
	}
	util.LogSuccess(stdout, "%supdated %s", drp, util.Bold("CHANGELOG.md"))

	// hooks
	if c.Hooks != nil && c.Hooks.BeforeStaging.Cmd != "" && allowHooks {
		util.LogExec(stdout, c.Hooks.BeforeStaging.Cmd, c.Hooks.BeforeStaging.Args)
		if !dryrun {
			cmd := exec.Command(c.Hooks.BeforeStaging.Cmd, c.Hooks.BeforeStaging.Args...)
			if r.Dir != "" {
				cmd.Dir = r.Dir
			}
			// FIXME: why isn't this working?
			// cmd.Env = append(cmd.Environ(), "ROAR_NEXT_VERSION="+next.String())
			out, err := cmd.Output()
			if err != nil {
				return "", err
			}
			util.LogExecOutput(stdout, string(out))
		}
	}

	// commit changes
	commitMsg := fmt.Sprintf("chore(release): release version %s", ntag)
	if !dryrun {
		// Since repository must be clean when running, we can just add all here
		if err := r.Add("."); err != nil {
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
	util.LogInfo(stdout, "please verify the applied changes and finalize the release by running\n\t%s", util.Bold("git push --follow-tags"))
	return ntag, nil
}

func updateFile(path, find, replace string, next semver.Version, epoch, drp string, dryrun bool, stdout io.Writer) error {
	// create replacement string: u.Replace is the template from the config
	replacement := strings.ReplaceAll(strings.ReplaceAll(replace, "{{version}}", next.String()), "{{epoch}}", epoch)
	// update file with replacement string
	if err := steps.FindAndReplace(path, find, replacement, dryrun); err != nil {
		return err
	}
	util.LogSuccess(stdout, "%supdated %s", drp, util.Bold(path))
	return nil
}
