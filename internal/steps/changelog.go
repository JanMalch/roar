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

	breakingCcs := make(map[string][]conventional.ConventionalCommit, 0)
	for _, k := range keys {
		ccs := ccLookup[k]
		for _, cc := range ccs {
			if cc.BreakingChange {
				breakingCcs[k] = append(breakingCcs[k], cc)
			}
		}
	}
	if len(breakingCcs) > 0 {
		sb.WriteString("### Breaking Changes\n\n")
		noScopeCcs := breakingCcs[""]
		if len(noScopeCcs) > 0 {
			for _, cc := range noScopeCcs {
				sb.WriteString(fmt.Sprintf("- %s\n", cc.BreakingChangeMessage))
			}
		}
		for k, ccs := range breakingCcs {
			if k == "" {
				continue
			}
			sb.WriteString(fmt.Sprintf("\n#### %s\n\n", k))
			for _, cc := range ccs {
				sb.WriteString(fmt.Sprintf("- %s\n", cc.BreakingChangeMessage))
			}
		}

		sb.WriteString("\n---\n\n")
	}

	hasNotableChanges := false
	for _, k := range keys {
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

		if k != "" {
			sb.WriteString(fmt.Sprintf("### %s\n\n", k))
		}

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

const upcomingSectionStartMarker = "<!-- ROAR:UPCOMING:START -->"
const upcomingSectionEndMarker = "<!-- ROAR:UPCOMING:END -->"

func removeUpcoming(content string) string {
	start := strings.Index(content, upcomingSectionStartMarker)
	if start == -1 {
		return content
	}
	end := strings.Index(content[start:], upcomingSectionEndMarker)
	if end == -1 {
		panic("upcoming section start marker without end marker")
	}
	return content[0:start] + content[start+end+len(upcomingSectionEndMarker):]
}

func generateUpcoming(version semver.Version, urlUpcoming string) string {
	if urlUpcoming == "" {
		return ""
	}
	return fmt.Sprintf(
		"%s\n[Upcoming Changes …](%s)\n%s",
		upcomingSectionStartMarker,
		strings.ReplaceAll(urlUpcoming, "{{version}}", version.String()),
		upcomingSectionEndMarker,
	)
}

func UpdateChangelog(
	path string,
	conf *models.ChangelogConfig,
	version semver.Version,
	prev *semver.Version,
	ccLookup map[string][]conventional.ConventionalCommit,
	today time.Time,
	dryrun bool,
) error {
	content := generateNewSection(conf, version, prev, ccLookup, today)

	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		if !dryrun {
			if conf.UrlUpcoming != "" {
				content = generateUpcoming(version, conf.UrlUpcoming) + "\n\n" + content
			}
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
		tail = strings.TrimPrefix(removeUpcoming(tail), "\n")
		if conf.UrlUpcoming != "" {
			content = generateUpcoming(version, conf.UrlUpcoming) + "\n\n" + content
		}
		return os.WriteFile(path, []byte(content+"\n\n"+tail), 0644)
	} else {
		return nil
	}

}
