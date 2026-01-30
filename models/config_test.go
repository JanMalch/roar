package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPatchChangelogEmptyUrlLeavesUnchanged(t *testing.T) {
	c := defaultConf.Changelog
	cc := patchChangelog(c, "")
	assert.Equal(t, c, cc)
}

func TestPatchChangelogGitLabSsh(t *testing.T) {
	c := defaultConf.Changelog
	cc := patchChangelog(c, "git@gitlab.ssh.example.com:group/subgroup/repository.git")
	assert.Equal(t, ChangelogConfig{
		Include:          c.Include,
		UrlCommit:        "https://gitlab.example.com/group/subgroup/repository/-/commit/{{hash}}",
		UrlBrowseAtTag:   "https://gitlab.example.com/group/subgroup/repository/-/tree/v{{version}}?ref_type=tags",
		UrlCompareTags:   "https://gitlab.example.com/group/subgroup/repository/-/compare/v{{previous}}...v{{version}}",
		UrlCommitsForTag: "https://gitlab.example.com/group/subgroup/repository/-/commits/v{{version}}?ref_type=tags",
		UrlUpcoming:      "https://gitlab.example.com/group/subgroup/repository/-/compare/v{{version}}...main",
	}, cc)
}

func TestPatchChangelogGitLabHttps(t *testing.T) {
	c := defaultConf.Changelog
	cc := patchChangelog(c, "https://gitlab.example.com/group/subgroup/repository.git")
	assert.Equal(t, ChangelogConfig{
		Include:          c.Include,
		UrlCommit:        "https://gitlab.example.com/group/subgroup/repository/-/commit/{{hash}}",
		UrlBrowseAtTag:   "https://gitlab.example.com/group/subgroup/repository/-/tree/v{{version}}?ref_type=tags",
		UrlCompareTags:   "https://gitlab.example.com/group/subgroup/repository/-/compare/v{{previous}}...v{{version}}",
		UrlCommitsForTag: "https://gitlab.example.com/group/subgroup/repository/-/commits/v{{version}}?ref_type=tags",
		UrlUpcoming:      "https://gitlab.example.com/group/subgroup/repository/-/compare/v{{version}}...main",
	}, cc)
}

func TestPatchChangelogGitHubSsh(t *testing.T) {
	c := defaultConf.Changelog
	cc := patchChangelog(c, "git@github.com:JanMalch/roar.git")
	assert.Equal(t, ChangelogConfig{
		Include:          c.Include,
		UrlCommit:        "https://github.com/JanMalch/roar/commit/{{hash}}",
		UrlBrowseAtTag:   "https://github.com/JanMalch/roar/tree/v{{version}}",
		UrlCompareTags:   "https://github.com/JanMalch/roar/compare/v{{previous}}...v{{version}}",
		UrlCommitsForTag: "https://github.com/JanMalch/roar/commits/v{{version}}",
		UrlUpcoming:      "https://github.com/JanMalch/roar/compare/v{{version}}...main",
	}, cc)
}

func TestPatchChangelogGitHubHttps(t *testing.T) {
	c := defaultConf.Changelog
	cc := patchChangelog(c, "https://github.com/JanMalch/roar.git")
	assert.Equal(t, ChangelogConfig{
		Include:          c.Include,
		UrlCommit:        "https://github.com/JanMalch/roar/commit/{{hash}}",
		UrlBrowseAtTag:   "https://github.com/JanMalch/roar/tree/v{{version}}",
		UrlCompareTags:   "https://github.com/JanMalch/roar/compare/v{{previous}}...v{{version}}",
		UrlCommitsForTag: "https://github.com/JanMalch/roar/commits/v{{version}}",
		UrlUpcoming:      "https://github.com/JanMalch/roar/compare/v{{version}}...main",
	}, cc)
}
