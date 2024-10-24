package steps

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/janmalch/roar/pkg/conventional"
)

func generateNewSection(version semver.Version, ccLookup map[string][]conventional.ConventionalCommit, includedTypes []string) string {
	keys := make([]string, 0, len(ccLookup))
	for k := range ccLookup {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## %s\n\n", version.String()))

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
			sb.WriteString("_No relevant changes._\n\n")
			continue
		}

		slices.SortFunc(relevantCcs, func(a, b conventional.ConventionalCommit) int {
			return strings.Compare(a.Type, b.Type)
		})

		sb.WriteString("| type | description | commit |\n")
		sb.WriteString("|---|---|---|\n")
		for _, c := range relevantCcs {
			sb.WriteString(fmt.Sprintf("| %s | %s | `%s` |\n", c.Type, c.Title, c.Hash[0:8]))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func UpdateChangelog(path string, version semver.Version, ccLookup map[string][]conventional.ConventionalCommit, includedTypes []string, dryrun bool) error {
	content := generateNewSection(version, ccLookup, includedTypes)

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
