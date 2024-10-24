package git

import (
	"bufio"
	"fmt"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type Repo struct {
	Dir string
}

type Commit struct {
	// The full hash of the commit
	Hash string
	// The first line of the commit message
	Message string
	// The date for the commit
	Date time.Time
}

func NewRepo(dir string) *Repo {
	return &Repo{Dir: dir}
}

func prepend(v string, tail []string) []string {
	s := make([]string, 1, 1+len(tail))
	s[0] = v
	s = append(s, tail...)
	return s
}

func (r *Repo) ExecGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if r.Dir != "" {
		cmd.Dir = r.Dir
	}
	out, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("failed command \"git %s\"", strings.Join(args, " ")))
	}
	return strings.TrimSpace(string(out)), err
}

func (r *Repo) PathOf(p string) string {
	return path.Join(r.Dir, p)
}

func (r *Repo) IsGitRepo() bool {
	out, _ := r.ExecGit("rev-parse", "--is-inside-work-tree")
	return out == "true"
}

func (r *Repo) IsClean() (bool, error) {
	out, err := r.ExecGit("status", "--porcelain")
	if err != nil {
		return false, err
	}
	return len(out) == 0, nil
}

func (r *Repo) CurrentBranchName() (string, error) {
	out, err := r.ExecGit("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return out, nil
}

func (r *Repo) GetOrigin() (string, error) {
	out, err := r.ExecGit("config", "--get", "remote.origin.url")
	// TODO: how to improve this?
	if err != nil {
		return "", nil
	}
	return out, nil
}

func (r *Repo) Pull() error {
	_, err := r.ExecGit("pull")
	return err
}

func (r *Repo) Add(pathspec ...string) error {
	_, err := r.ExecGit(prepend("add", pathspec)...)
	return err
}

func (r *Repo) Commit(message string) error {
	_, err := r.ExecGit("commit", "-m", message)
	return err
}

func (r *Repo) AddTag(name string) error {
	_, err := r.ExecGit("tag", name)
	return err
}

func (r *Repo) LatestVersionTag() (string, error) {
	cmd := exec.Command("git", "tag", "--sort=-version:refname")
	if r.Dir != "" {
		cmd.Dir = r.Dir
	}
	bout, err := cmd.CombinedOutput()
	out := strings.TrimSpace(string(bout))
	if err != nil {
		if strings.Contains(out, "No names found") {
			return "", nil
		}
		return "", errors.Wrap(err, fmt.Sprintln("failed command \"git describe --tags --abbrev=0\""))
	}
	for _, line := range strings.Split(out, "\n") {
		tline := strings.TrimSpace(line)
		if strings.HasPrefix(tline, "v") {
			return tline, nil
		}
	}
	return out, nil
}

func (r *Repo) CommitLogSince(tag string) ([]Commit, error) {
	var out string
	var err error
	if tag != "" {
		out, err = r.ExecGit("log", "--pretty=%at %H %s", "HEAD..."+tag)
	} else {
		out, err = r.ExecGit("log", "--pretty=%at %H %s")
	}
	if err != nil {
		return nil, err
	}
	commits := make([]Commit, 0)

	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 3)
		at, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return nil, err
		}
		commits = append(commits, Commit{
			Date:    time.Unix(at, 0),
			Hash:    parts[1],
			Message: parts[2],
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return commits, nil
}
