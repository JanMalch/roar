package git

import (
	"bufio"
	"bytes"
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
	// The full message of the commit
	Message string
	// The date for the commit
	Date time.Time
}

// The first line of the commit message
func (c Commit) Subject() string {
	i := strings.IndexByte(c.Message, '\n')
	if i < 0 {
		return c.Message
	}
	return strings.TrimSpace(c.Message[0:i])
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

func (r *Repo) HasCommits() (bool, error) {
	out, err := r.ExecGit("rev-list", "-n1", "--all")
	if err != nil {
		return false, err
	}
	return len(out) > 0, nil
}

func (r *Repo) CurrentBranchName() (string, error) {
	out, err := r.ExecGit("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return out, nil
}

func (r *Repo) OriginUrl() (string, error) {
	out, err := r.ExecGit("remote", "get-url", "origin")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func (r *Repo) Add(pathspec ...string) error {
	_, err := r.ExecGit(prepend("add", pathspec)...)
	return err
}

func (r *Repo) ListModified() ([]string, error) {
	out, err := r.ExecGit("ls-files", "--modified")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.ReplaceAll(strings.TrimSpace(out), "\r\n", "\n"), "\n")
	result := make([]string, 0)
	for _, line := range lines {
		tl := strings.TrimSpace(line)
		if tl != "" {
			result = append(result, tl)
		}
	}
	return result, nil
}

func (r *Repo) Commit(message string) error {
	_, err := r.ExecGit("commit", "-m", message)
	return err
}

func (r *Repo) FetchTags() error {
	_, err := r.ExecGit("fetch", "--tags")
	return err
}

func (r *Repo) AddTag(name string) error {
	_, err := r.ExecGit("tag", "-a", name, "-m", "Release "+name)
	return err
}

func (r *Repo) LatestVersionTag() (string, error) {
	// not using r.ExecGit because of CombinedOutput
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
	cmd := exec.Command("git", "log", "--pretty=format:%at%x1f%H%x1f%B%x1e")
	if tag != "" {
		cmd.Args = append(cmd.Args, "HEAD..."+tag)
	}
	if r.Dir != "" {
		cmd.Dir = r.Dir
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanBytes)

	commits := make([]Commit, 0)

	var buf bytes.Buffer
	for scanner.Scan() {
		b := scanner.Bytes()[0]
		if b == 0x1e { // record separator
			parts := strings.SplitN(buf.String(), "\x1f", 3)
			at, _ := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
			commits = append(commits, Commit{
				Date:    time.Unix(at, 0),
				Hash:    parts[1],
				Message: strings.TrimSpace(parts[2]),
			})
			buf.Reset()
		} else {
			buf.WriteByte(b)
		}
	}

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return commits, nil
}
