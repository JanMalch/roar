package conventional_test

import (
	"strconv"
	"testing"

	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/janmalch/roar/pkg/conventional"
	"github.com/janmalch/roar/util"
	"github.com/stretchr/testify/assert"
)

func TestParseFeatureNoScope(t *testing.T) {
	c := &object.Commit{
		Hash:    plumbing.NewHash("abc123"),
		Message: "feat: test commit",
	}
	cc := conventional.Parse(c)
	if assert.NotNil(t, cc) {
		assert.Equal(t, "feat", cc.Type)
		assert.Equal(t, util.MINOR_CHANGE, cc.Change)
		assert.Equal(t, "", cc.Scope)
		assert.Equal(t, "test commit", cc.Title)
		assert.False(t, cc.BreakingChange, "'%s' should not indicate a breaking change", c.Message)
	}
}

func TestParseBreakingChangeMessage(t *testing.T) {
	c := &object.Commit{
		Hash: plumbing.NewHash("abc123"),
		Message: `fix(users): test fix
		
Here be dragons.

BREAKING CHANGE: This breaks the dragon.

Refs: #12345`,
	}
	cc := conventional.Parse(c)
	if assert.NotNil(t, cc) {
		assert.Equal(t, "fix", cc.Type)
		assert.Equal(t, util.MAJOR_CHANGE, cc.Change)
		assert.Equal(t, "users", cc.Scope)
		assert.Equal(t, "test fix", cc.Title)
		assert.True(t, cc.BreakingChange)
		assert.Equal(t, "This breaks the dragon.", cc.BreakingChangeMessage)
	}
}

func TestParseFixWithScope(t *testing.T) {
	c := &object.Commit{
		Hash:    plumbing.NewHash("abc123"),
		Message: "fix(users): test fix",
	}
	cc := conventional.Parse(c)
	if assert.NotNil(t, cc) {
		assert.Equal(t, "fix", cc.Type)
		assert.Equal(t, util.PATCH_CHANGE, cc.Change)
		assert.Equal(t, "users", cc.Scope)
		assert.Equal(t, "test fix", cc.Title)
		assert.False(t, cc.BreakingChange, "'%s' should not indicate a breaking change", c.Message)
	}
}

func TestParseBreakingChance(t *testing.T) {
	c := &object.Commit{
		Hash:    plumbing.NewHash("abc123"),
		Message: "fix(users)!: test fix",
	}
	cc := conventional.Parse(c)
	if assert.NotNil(t, cc) {
		assert.Equal(t, "fix", cc.Type)
		assert.Equal(t, util.MAJOR_CHANGE, cc.Change)
		assert.Equal(t, "users", cc.Scope)
		assert.Equal(t, "test fix", cc.Title)
		assert.Equal(t, "test fix", cc.BreakingChangeMessage)
		assert.True(t, cc.BreakingChange, "'%s' should indicate a breaking change", c.Message)
	}
}

func createLog(msgs ...string) []*object.Commit {
	log := make([]*object.Commit, 0, len(msgs))

	for i, msg := range msgs {
		log = append(log, &object.Commit{
			Hash:    plumbing.NewHash("abc" + strconv.Itoa(i)),
			Message: msg,
		})
	}

	return log
}
