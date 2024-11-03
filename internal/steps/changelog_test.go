package steps

import (
	"testing"
	"time"

	"github.com/Masterminds/semver"
	"github.com/janmalch/roar/pkg/conventional"
	"github.com/janmalch/roar/pkg/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateNewSectionWithMixOfTypesAndScopes(t *testing.T) {
	latest, err := semver.NewVersion("0.1.0")
	require.NoError(t, err)
	gs, _ := NewGitService("", "")
	actual := generateNewSection(gs, *latest, nil, map[string][]conventional.ConventionalCommit{
		"users": {
			*conventional.Parse(git.Commit{
				Message: "fix(users): add user fix",
				Hash:    "00000001",
				Date:    time.Now(),
			}),
			*conventional.Parse(git.Commit{
				Message: "feat(users): add user feature",
				Hash:    "00000002",
				Date:    time.Now(),
			})},
		"": {
			*conventional.Parse(git.Commit{
				Message: "fix: add some fix",
				Hash:    "00000003",
				Date:    time.Now(),
			}),
			*conventional.Parse(git.Commit{
				Message: "feat: add some feature",
				Hash:    "00000004",
				Date:    time.Now(),
			}),
		},
	}, []string{"feat", "fix", "refactor"},
	)
	expected := `## 0.1.0

| type | description | commit |
|---|---|---|
| feat | add some feature | ` + "`" + `00000004` + "`" + ` |
| fix | add some fix | ` + "`" + `00000003` + "`" + ` |

### users

| type | description | commit |
|---|---|---|
| feat | add user feature | ` + "`" + `00000002` + "`" + ` |
| fix | add user fix | ` + "`" + `00000001` + "`" + ` |

`
	assert.Equal(t, expected, actual)

}
