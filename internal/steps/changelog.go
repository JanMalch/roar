package steps

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/janmalch/roar/pkg/conventional"
)

func generateNewSection(gs GitService, version semver.Version, prev *semver.Version, ccLookup map[string][]conventional.ConventionalCommit, includedTypes []string, today time.Time) string {
	keys := make([]string, 0, len(ccLookup))
	for k := range ccLookup {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## %s - %s\n\n", gs.FmtHeader(version.String()), today.Format("January 2, 2006")))

	cmp := ""
	if prev != nil {
		cmp = gs.FmtDiff(prev.String(), version.String())
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
			if slices.Contains(includedTypes, cc.Type) {
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
			sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", c.Type, c.Title, gs.FmtCommit(c.Hash)))
		}
		sb.WriteString("\n")
	}

	if !hasNotableChanges {
		sb.WriteString("_No notable changes._\n\n")
	}

	return sb.String()
}

func UpdateChangelog(path string, gs GitService, version semver.Version, prev *semver.Version, ccLookup map[string][]conventional.ConventionalCommit, includedTypes []string, today time.Time, dryrun bool) error {
	content := generateNewSection(gs, version, prev, ccLookup, includedTypes, today)

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
