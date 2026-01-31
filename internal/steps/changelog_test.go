package steps

import (
	"testing"
	"time"

	"github.com/Masterminds/semver"
	"github.com/janmalch/roar/models"
	"github.com/janmalch/roar/pkg/conventional"
	"github.com/janmalch/roar/pkg/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateUpcoming(t *testing.T) {
	latest, err := semver.NewVersion("0.1.0")
	require.NoError(t, err)
	actual := generateUpcoming(*latest, "https://github.com/JanMalch/roar/compare/v{{version}}...main")
	expected := `<!-- ROAR:UPCOMING:START -->
[Upcoming Changes …](https://github.com/JanMalch/roar/compare/v0.1.0...main)
<!-- ROAR:UPCOMING:END -->`
	assert.Equal(t, expected, actual)
}

func TestRemoveUpcomingForNonExistent(t *testing.T) {
	assert.Equal(t, "Hello", removeUpcoming("Hello"))
}

func TestRemoveUpcomingForExistent(t *testing.T) {
	actual := removeUpcoming(`What

<!-- ROAR:UPCOMING:START -->
[Upcoming  Changes …](https://github.com/JanMalch/roar/compare/v0.1.0...main)
<!-- ROAR:UPCOMING:END -->

Hello`)
	expected := `What



Hello`
	assert.Equal(t, expected, actual)
}

func TestGenerateNewSectionWithMixOfTypesAndScopes(t *testing.T) {
	latest, err := semver.NewVersion("0.1.0")
	require.NoError(t, err)
	conf := models.ChangelogConfig{
		Include:          []string{"feat", "fix", "refactor"},
		UrlCommit:        "",
		UrlBrowseAtTag:   "",
		UrlCompareTags:   "",
		UrlCommitsForTag: "",
		UrlUpcoming:      "",
	}
	today := time.Date(2024, time.November, 8, 12, 0, 0, 0, time.UTC)
	actual := generateNewSection(&conf, *latest, nil, map[string][]conventional.ConventionalCommit{
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
	}, today,
	)
	expected := `## 0.1.0 - November 8, 2024

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

func TestSkipEmptySections(t *testing.T) {
	latest, err := semver.NewVersion("0.1.0")
	require.NoError(t, err)
	conf := models.ChangelogConfig{
		Include:          []string{"feat", "fix", "refactor"},
		UrlCommit:        "",
		UrlBrowseAtTag:   "",
		UrlCompareTags:   "",
		UrlCommitsForTag: "",
	}
	today := time.Date(2024, time.November, 8, 12, 0, 0, 0, time.UTC)
	actual := generateNewSection(&conf, *latest, nil, map[string][]conventional.ConventionalCommit{
		"deps": {
			*conventional.Parse(git.Commit{
				Message: "chore(deps): update dependencies",
				Hash:    "00000001",
				Date:    time.Now(),
			}),
		},
		"": {
			*conventional.Parse(git.Commit{
				Message: "chore: foo bar",
				Hash:    "00000002",
				Date:    time.Now(),
			}),
		},
	}, today,
	)
	expected := `## 0.1.0 - November 8, 2024

_No notable changes._

`
	assert.Equal(t, expected, actual)

}

func TestBreakingChanges(t *testing.T) {
	latest, err := semver.NewVersion("0.1.0")
	require.NoError(t, err)
	conf := models.ChangelogConfig{
		Include:          []string{"feat", "fix", "refactor"},
		UrlCommit:        "",
		UrlBrowseAtTag:   "",
		UrlCompareTags:   "",
		UrlCommitsForTag: "",
		UrlUpcoming:      "",
	}
	today := time.Date(2024, time.November, 8, 12, 0, 0, 0, time.UTC)
	actual := generateNewSection(&conf, *latest, nil, map[string][]conventional.ConventionalCommit{
		"user": {
			*conventional.Parse(git.Commit{
				Message: "feat(user)!: seems broken",
				Hash:    "00000001",
				Date:    time.Now(),
			}),
		},
		"": {
			*conventional.Parse(git.Commit{
				Message: `fix: foo bar
				
BREAKING CHANGE: Well, there it is.`,
				Hash: "00000002",
				Date: time.Now(),
			}),
		},
	}, today,
	)
	expected := `## 0.1.0 - November 8, 2024

### Breaking Changes

- Well, there it is.

#### user

- seems broken

---

| type | description | commit |
|---|---|---|
| fix | foo bar | ` + "`" + `00000002` + "`" + ` |

### user

| type | description | commit |
|---|---|---|
| feat | seems broken | ` + "`" + `00000001` + "`" + ` |

`
	assert.Equal(t, expected, actual)

}
