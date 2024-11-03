package steps

import (
	"fmt"
	"strings"
)

type GitService interface {
	FmtCommit(hash string) string
	FmtHeader(version string) string
	FmtDiff(from, to string) string
}

func DetectGitService(origin, url, service string) (GitService, bool) {
	u := url
	if u == "" {
		u = strings.TrimSuffix(origin, ".git")
	}
	s := service
	if s == "" {
		lu := strings.ToLower(u)
		if strings.Contains(lu, "github") {
			s = "github"
		} else if strings.Contains(lu, "gitlab") {
			s = "gitlab"
		} else if strings.Contains(lu, "bitbucket") {
			s = "bitbucket"
		}
	}
	return NewGitService(u, s)
}

func NewGitService(url, service string) (GitService, bool) {
	switch strings.ToLower(service) {
	case "github":
		return &github{base: url}, true
	case "gitlab":
		return &gitlab{base: url}, true
	case "bitbucket":
		return &bitbucket{base: url}, true
	}
	return &unknown{}, false
}

type unknown struct {
}

func (u *unknown) FmtCommit(hash string) string    { return fmt.Sprintf("`%s`", hash[0:8]) }
func (u *unknown) FmtHeader(version string) string { return version }
func (u *unknown) FmtDiff(from, to string) string  { return "" }

type github struct {
	base string
}

func (gh *github) FmtCommit(hash string) string {
	return fmt.Sprintf("[`%s`](%s/commit/%s)", hash[0:8], gh.base, hash)
}
func (gh *github) FmtHeader(version string) string {
	return fmt.Sprintf("[%s](%s/tree/v%s)", version, gh.base, version)
}
func (gh *github) FmtDiff(from, to string) string {
	return fmt.Sprintf("**Full Changelog:** [`v%s...v%s`](%s/compare/v%s...v%s)", from, to, gh.base, from, to)
}

type gitlab struct {
	base string
}

func (gl *gitlab) FmtCommit(hash string) string {
	return fmt.Sprintf("[`%s`](%s/-/commit/%s)", hash[0:8], gl.base, hash)
}
func (gl *gitlab) FmtHeader(version string) string {
	return fmt.Sprintf("[%s](%s/-/tree/v%s?ref_type=versions)", version, gl.base, version)
}
func (gl *gitlab) FmtDiff(from, to string) string {
	return fmt.Sprintf("**Full Changelog:** [`v%s...v%s`](%s/-/compare?from=v%s&t=v%s)", from, to, gl.base, from, to)
}

type bitbucket struct {
	base string
}

func (bb *bitbucket) FmtCommit(hash string) string {
	return fmt.Sprintf("[`%s`](%s/commits/%s)", hash[0:8], bb.base, hash)
}
func (bb *bitbucket) FmtHeader(version string) string {
	return fmt.Sprintf("[%s](%s/src/%s/)", version, bb.base, version)
}
func (bb *bitbucket) FmtDiff(from, to string) string {
	return fmt.Sprintf("**Full Changelog:** [`v%s...v%s`](%s/compare/v%s..v%s)", from, to, bb.base, from, to)
}
