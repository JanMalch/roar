package steps

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/janmalch/roar/models"
	"github.com/janmalch/roar/pkg/conventional"
)

func generateNewSection(conf *models.ChangelogConfig, version semver.Version, prev *semver.Version, ccLookup map[string][]conventional.ConventionalCommit, today time.Time) string {
	keys := make([]string, 0, len(ccLookup))
	for k := range ccLookup {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	fmtToday := today.Format("January 2, 2006")

	var sb strings.Builder
	if conf.UrlBrowseAtTag != "" {
		sb.WriteString(fmt.Sprintf("## [%s](%s) - %s\n\n", version.String(), strings.NewReplacer("{{version}}", version.String()).Replace(conf.UrlBrowseAtTag), fmtToday))
	} else {
		sb.WriteString(fmt.Sprintf("## %s - %s\n\n", version.String(), fmtToday))
	}

	cmp := ""
	if prev != nil {
		url := strings.NewReplacer("{{previous}}", prev.String(), "{{version}}", version.String()).Replace(conf.UrlCompareTags)
		if url != "" {
			cmp = fmt.Sprintf("**Full Changelog:** [`v%s...v%s`](%s)", prev.String(), version.String(), url)
		}
	} else {
		url := strings.NewReplacer("{{version}}", version.String()).Replace(conf.UrlCommitsForTag)
		if url != "" {
			cmp = fmt.Sprintf("**Full Changelog:** [`v%s`](%s)", version.String(), url)
		}
	}
	if cmp != "" {
		sb.WriteString(cmp + "\n\n")
	}

	hasNotableChanges := false
	for _, k := range keys {
		if k != "" {
			sb.WriteString(fmt.Sprintf("### %s\n\n", k))
		}
		ccs := ccLookup[k]
		relevantCcs := make([]conventional.ConventionalCommit, 0)

		for _, cc := range ccs {
			if slices.Contains(conf.Include, cc.Type) {
				relevantCcs = append(relevantCcs, cc)
			}
		}

		if len(relevantCcs) == 0 {
			continue
		}
		hasNotableChanges = true

		slices.SortFunc(relevantCcs, func(a, b conventional.ConventionalCommit) int {
			return strings.Compare(a.Type, b.Type)
		})

		sb.WriteString("| type | description | commit |\n")
		sb.WriteString("|---|---|---|\n")
		for _, c := range relevantCcs {
			fmtCommit := ""
			if conf.UrlCommit != "" {
				fmtCommit = fmt.Sprintf("[`%s`](%s)", c.Hash[0:8], strings.NewReplacer("{{hash}}", c.Hash).Replace(conf.UrlCommit))
			} else {
				fmtCommit = fmt.Sprintf("`%s`", c.Hash[0:8])
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", c.Type, c.Title, fmtCommit))
		}
		sb.WriteString("\n")
	}

	if !hasNotableChanges {
		sb.WriteString("_No notable changes._\n\n")
	}

	return sb.String()
}

func UpdateChangelog(path string, conf *models.ChangelogConfig, version semver.Version, prev *semver.Version, ccLookup map[string][]conventional.ConventionalCommit, today time.Time, dryrun bool) error {
	content := generateNewSection(conf, version, prev, ccLookup, today)

	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		if !dryrun {
			return os.WriteFile(path, []byte(content), 0644)
		} else {
			return nil
		}
	}
	if err != nil {
		return err
	}
	tail := string(b)

	if !dryrun {
		return os.WriteFile(path, []byte(content+"\n\n"+tail), 0644)
	} else {
		return nil
	}

}
