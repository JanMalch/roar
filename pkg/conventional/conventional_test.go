package conventional_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/janmalch/roar/pkg/conventional"
	"github.com/janmalch/roar/pkg/git"
	"github.com/janmalch/roar/util"
	"github.com/stretchr/testify/assert"
)

func TestParseFeatureNoScope(t *testing.T) {
	c := git.Commit{
		Message: "feat: test commit",
		Hash:    "H123",
		Date:    time.Now(),
	}
	cc := conventional.Parse(c)
	if assert.NotNil(t, cc) {
		assert.Equal(t, c.Date, cc.Date)
		assert.Equal(t, c.Hash, cc.Hash)
		assert.Equal(t, c.Message, cc.Message)
		assert.Equal(t, "feat", cc.Type)
		assert.Equal(t, util.MINOR_CHANGE, cc.Change)
		assert.Equal(t, "", cc.Scope)
		assert.Equal(t, "test commit", cc.Title)
		assert.False(t, cc.BreakingChange, "'%s' should not indicate a breaking change", c.Message)
	}
}

func TestParseFixWithScope(t *testing.T) {
	c := git.Commit{
		Message: "fix(users): test fix",
		Hash:    "H123",
		Date:    time.Now(),
	}
	cc := conventional.Parse(c)
	if assert.NotNil(t, cc) {
		assert.Equal(t, c.Date, cc.Date)
		assert.Equal(t, c.Hash, cc.Hash)
		assert.Equal(t, c.Message, cc.Message)
		assert.Equal(t, "fix", cc.Type)
		assert.Equal(t, util.PATCH_CHANGE, cc.Change)
		assert.Equal(t, "users", cc.Scope)
		assert.Equal(t, "test fix", cc.Title)
		assert.False(t, cc.BreakingChange, "'%s' should not indicate a breaking change", c.Message)
	}
}

func TestParseBreakingChance(t *testing.T) {
	c := git.Commit{
		Message: "fix(users)!: test fix",
		Hash:    "H123",
		Date:    time.Now(),
	}
	cc := conventional.Parse(c)
	if assert.NotNil(t, cc) {
		assert.Equal(t, c.Date, cc.Date)
		assert.Equal(t, c.Hash, cc.Hash)
		assert.Equal(t, c.Message, cc.Message)
		assert.Equal(t, "fix", cc.Type)
		assert.Equal(t, util.MAJOR_CHANGE, cc.Change)
		assert.Equal(t, "users", cc.Scope)
		assert.Equal(t, "test fix", cc.Title)
		assert.True(t, cc.BreakingChange, "'%s' should indicate a breaking change", c.Message)
	}
}

func createLog(msgs ...string) []git.Commit {
	log := make([]git.Commit, 0, len(msgs))

	for i, msg := range msgs {
		log = append(log, git.Commit{
			Message: msg,
			Hash:    strconv.Itoa(i),
			Date:    time.Now(),
		})
	}

	return log
}

func TestCollectConventionalCommitsMinor(t *testing.T) {
	log := createLog(
		"fix(users): test fix",
		"feat(users): test feat",
		"fix(auth): fix auth this time")

	lookup, change, err := conventional.Collect(log)
	if assert.Nil(t, err) {
		assert.Equal(t, util.MINOR_CHANGE, change)
		assert.Equal(t, 2, len(lookup))
		assert.Equal(t, 2, len(lookup["users"]))
		assert.Equal(t, 1, len(lookup["auth"]))
	}
}
