package models

import (
	"bytes"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type UpdateConfig struct {
	File    string `toml:"file"`
	Find    string `toml:"find"`
	Replace string `toml:"replace"`
}

type ChangelogConfig struct {
	Include          []string `toml:"include"`
	UrlCommit        string   `toml:"url_commit"`
	UrlBrowseAtTag   string   `toml:"url_browse_at_tag"`
	UrlCompareTags   string   `toml:"url_compare_tags"`
	UrlCommitsForTag string   `toml:"url_commits_for_tag"`
}

type Config struct {
	Branch    string          `toml:"branch"`
	Updates   []UpdateConfig  `toml:"update"`
	Changelog ChangelogConfig `toml:"changelog"`
}

var defaultConf = &Config{
	Updates: []UpdateConfig{},
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

var changelogUrlNote = `# FIXME: Remove this line, after you verified the entire configuration.
`

var (
	ErrFindIsEmpty    = errors.New("\"find\" may not be empty in config")
	ErrReplaceIsEmpty = errors.New("\"replace\" may not be empty in config")
	ErrFixmesInToml   = errors.New("found FIXME comments in the configuration file")
)

func patchChangelogForGitHub(c ChangelogConfig, gitHubBase string) ChangelogConfig {
	c.UrlCommit = gitHubBase + "/commit/{{hash}}"
	c.UrlBrowseAtTag = gitHubBase + "/tree/v{{version}}"
	c.UrlCompareTags = gitHubBase + "/compare/v{{previous}}...v{{version}}"
	c.UrlCommitsForTag = gitHubBase + "/commits/v{{version}}"
	return c
}

func patchChangelogForGitLab(c ChangelogConfig, gitLabBase string) ChangelogConfig {
	c.UrlCommit = gitLabBase + "/-/commit/{{hash}}"
	c.UrlBrowseAtTag = gitLabBase + "/-/tree/v{{version}}?ref_type=tags"
	c.UrlCompareTags = gitLabBase + "/-/compare/v{{previous}}...v{{version}}"
	c.UrlCommitsForTag = gitLabBase + "/-/commits/v{{version}}?ref_type=tags"
	return c
}

func patchChangelog(c ChangelogConfig, gitRemoteUrl string) ChangelogConfig {
	if gitRemoteUrl == "" {
		return c
	}
	if strings.HasPrefix(gitRemoteUrl, "git@") {
		parsed, err := url.Parse("https://" + strings.ReplaceAll(strings.TrimSuffix(gitRemoteUrl[4:], ".git"), ":", "/"))
		if err != nil {
			return c
		}
		if strings.Contains(gitRemoteUrl, "gitlab") {
			base := strings.Replace(parsed.String(), ".ssh.", ".", 1)
			return patchChangelogForGitLab(c, base)
		}
		if strings.Contains(gitRemoteUrl, "github") {
			return patchChangelogForGitHub(c, parsed.String())
		}
	}
	parsed, err := url.Parse(gitRemoteUrl)
	if err != nil || parsed.Scheme != "https" {
		return c
	}
	parsed.User = nil
	parsed.Host = parsed.Hostname() // remove port, just in case
	base := strings.TrimSuffix(parsed.String(), ".git")
	if strings.Contains(gitRemoteUrl, "gitlab") {
		return patchChangelogForGitLab(c, base)
	}
	if strings.Contains(gitRemoteUrl, "github") {
		return patchChangelogForGitHub(c, base)
	}
	return c
}

func createDefaultUpdates(configFilePath string) []UpdateConfig {
	dir := filepath.Dir(configFilePath)
	if _, err := os.Stat(filepath.Join(dir, "gradle.properties")); errors.Is(err, os.ErrNotExist) {
		return []UpdateConfig{
			{
				File:    "gradle.properties",
				Find:    "VERSION_NAME=",
				Replace: "VERSION_NAME={{version}}",
			},
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "openapi.yml")); errors.Is(err, os.ErrNotExist) {
		return []UpdateConfig{
			{
				File:    "openapi.yml",
				Find:    "  version: ",
				Replace: "  version: {{version}}",
			},
		}
	}

	// have anything as default
	return []UpdateConfig{
		{
			File:    "openapi.yaml",
			Find:    "  version: ",
			Replace: "  version: {{version}}",
		},
	}
}

// Returns a config and a bool, if the returned config is newly created.
func ConfigFromFile(path string, gitRemoteUrl string) (*Config, bool, error) {
	rawBytes, err := os.ReadFile(path) // just pass the file name
	if err != nil && !os.IsNotExist(err) {
		return nil, false, err
	}
	rawContent := string(rawBytes)
	if strings.Contains(rawContent, "# FIXME") {
		return nil, false, ErrFixmesInToml
	}

	var conf Config
	_, err = toml.DecodeFile(path, &conf)
	if err != nil && !os.IsNotExist(err) {
		return nil, false, err
	}
	if os.IsNotExist(err) {
		newConf := *defaultConf
		newConf.Updates = createDefaultUpdates(path)
		newConf.Changelog = patchChangelog(newConf.Changelog, gitRemoteUrl)
		buff := new(bytes.Buffer)
		enc := toml.NewEncoder(buff)
		enc.Indent = ""
		if err := enc.Encode(newConf); err != nil {
			return nil, false, err
		}
		c := headerText + buff.String() + changelogUrlNote
		if err := os.WriteFile(path, []byte(c), 0644); err != nil {
			return nil, false, err
		}
		return &newConf, true, nil
	}

	if len(conf.Changelog.Include) == 0 {
		conf.Changelog.Include = defaultConf.Changelog.Include
	}
	for _, u := range conf.Updates {
		if len(strings.TrimSpace(u.Find)) == 0 {
			return nil, false, ErrFindIsEmpty
		}
		if len(strings.TrimSpace(u.Replace)) == 0 {
			return nil, false, ErrReplaceIsEmpty
		}
	}
	return &conf, false, nil
}
