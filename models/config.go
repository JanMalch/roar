package models

import (
	"bytes"
	"errors"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

type ChangelogConfig struct {
	Include          []string `toml:"include"`
	UrlCommit        string   `toml:"url_commit"`
	UrlBrowseAtTag   string   `toml:"url_browse_at_tag"`
	UrlCompareTags   string   `toml:"url_compare_tags"`
	UrlCommitsForTag string   `toml:"url_commits_for_tag"`
}

type Config struct {
	File      string          `toml:"file"`
	Find      string          `toml:"find"`
	Replace   string          `toml:"replace"`
	Changelog ChangelogConfig `toml:"changelog"`
}

var defaultConf = Config{
	File:    "openapi.yml",
	Find:    "  version: ",
	Replace: "  version: {{version}}",
	Changelog: ChangelogConfig{
		Include:          []string{"feat", "fix", "refactor"},
		UrlCommit:        "https://github.com/owner/repo/commit/{{hash}}",
		UrlBrowseAtTag:   "https://github.com/owner/repo/tree/v{{version}}",
		UrlCompareTags:   "https://github.com/owner/repo/compare/v{{previous}}...v{{version}}",
		UrlCommitsForTag: "https://github.com/owner/repo/commits/v{{version}}",
	},
}

var headerText = `# Configuration for the roar CLI
# https://github.com/JanMalch/roar

`

var changelogUrlNote = `# FIXME: change the changelog URL templates to point to your repository. Using GitHub is just an example.
`

var (
	ErrFindIsEmpty          = errors.New("\"find\" may not be empty in config")
	ErrReplaceIsEmpty       = errors.New("\"replace\" may not be empty in config")
	ErrDefaultChangelogUrls = errors.New("URLs in \"changelog\" may not be the exemplary default")
)

// Returns a config and a bool, if the returned config is newly created.
func ConfigFromFile(path string) (*Config, bool, error) {
	var conf Config
	_, err := toml.DecodeFile(path, &conf)

	if err != nil && !os.IsNotExist(err) {
		return nil, false, err
	}
	if os.IsNotExist(err) {
		buff := new(bytes.Buffer)
		enc := toml.NewEncoder(buff)
		enc.Indent = ""
		if err := enc.Encode(defaultConf); err != nil {
			return nil, false, err
		}
		c := headerText + buff.String() + changelogUrlNote
		if err := os.WriteFile(path, []byte(c), 0644); err != nil {
			return nil, false, err
		}
		return &defaultConf, true, nil
	}

	if len(conf.Changelog.Include) == 0 {
		conf.Changelog.Include = defaultConf.Changelog.Include
	}
	if len(strings.TrimSpace(conf.Find)) == 0 {
		return nil, false, ErrFindIsEmpty
	}
	if len(strings.TrimSpace(conf.Replace)) == 0 {
		return nil, false, ErrReplaceIsEmpty
	}
	if strings.HasPrefix(conf.Changelog.UrlCommit, "https://github.com/owner/repo/") {
		return nil, false, ErrDefaultChangelogUrls
	}
	if strings.HasPrefix(conf.Changelog.UrlBrowseAtTag, "https://github.com/owner/repo/") {
		return nil, false, ErrDefaultChangelogUrls
	}
	if strings.HasPrefix(conf.Changelog.UrlCompareTags, "https://github.com/owner/repo/") {
		return nil, false, ErrDefaultChangelogUrls
	}
	if strings.HasPrefix(conf.Changelog.UrlCommitsForTag, "https://github.com/owner/repo/") {
		return nil, false, ErrDefaultChangelogUrls
	}
	return &conf, false, nil
}
