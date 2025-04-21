package steps

import (
	"os"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/janmalch/roar/pkg/git"
)

var (
	ErrInputDoesNotExist     = errors.New("input file does not exist")
	ErrNotAGitRepo           = errors.New("directory is not a git repository")
	ErrRepoNotClean          = errors.New("git repository is not clean")
	ErrInvalidFind           = errors.New("find argument is invalid")
	ErrReplaceNoPlaceholders = errors.New("replace argument contains no valid placeholders")
	ErrInvalidBranch         = errors.New("current branch doesn't match expected branch")
)

func ValidateFind(find string) error {
	if find == "" {
		return ErrInvalidFind
	}
	return nil
}

func ValidateReplace(replace string) error {
	if !strings.Contains(replace, "{{version}}") && !strings.Contains(replace, "{{epoch}}") {
		return ErrReplaceNoPlaceholders
	}
	return nil
}

func ConfirmInputExists(input string) error {
	if _, err := os.Stat(input); errors.Is(err, os.ErrNotExist) {
		return ErrInputDoesNotExist
	} else if err != nil {
		return errors.Wrap(err, "failed to check if input file exists")
	} else {
		return nil
	}
}

func ConfirmGitRepo(r *git.Repo) error {
	if !r.IsGitRepo() {
		return ErrNotAGitRepo
	} else {
		return nil
	}
}

func ValidateBranch(r *git.Repo, expected string) (string, error) {
	branch, err := r.CurrentBranchName()
	if err != nil {
		return branch, errors.Wrap(err, "failed to determine current branch name")
	}
	if expected == "" {
		return branch, nil
	}
	if strings.HasPrefix(expected, "^") || strings.HasPrefix(expected, "(?i)^") {
		regex := regexp.MustCompile(expected)
		if regex.MatchString(branch) {
			return branch, nil
		} else {
			return branch, ErrInvalidBranch
		}
	} else {
		if expected == branch {
			return branch, nil
		} else {
			return branch, ErrInvalidBranch
		}
	}
}

func ConfirmClean(r *git.Repo) error {
	if clean, err := r.IsClean(); err != nil {
		return errors.Wrap(err, "failed to check repository status")
	} else if !clean {
		return ErrRepoNotClean
	}
	return nil
}
