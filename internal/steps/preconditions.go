package steps

import (
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/janmalch/roar/pkg/git"
)

var (
	ErrInputDoesNotExist = errors.New("input file does not exist")
	ErrNotAGitRepo       = errors.New("directory is not a git repository")
	ErrRepoNotClean      = errors.New("git repository is not clean")
	ErrInvalidFind       = errors.New("find argument is invalid")
	ErrInvalidReplace    = errors.New("replace argument is invalid")
)

func ValidateFind(find string) error {
	if find == "" {
		return ErrInvalidFind
	}
	return nil
}

func ValidateReplace(replace string) error {
	if !strings.Contains(replace, "{{version}}") {
		return ErrInvalidReplace
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

func DetermineCurrentBranch(r *git.Repo) (string, error) {
	if branch, err := r.CurrentBranchName(); err != nil {
		return "", errors.Wrap(err, "failed to determine current branch name")
	} else {
		return branch, nil
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
