package run

import (
	"testing"

	"github.com/janmalch/roar/models"
	"github.com/stretchr/testify/assert"
)

func TestPatch(t *testing.T) {
	conf := &models.Config{
		Changelog: models.ChangelogConfig{
			Include: []string{"feat", "fix"},
		},
	}
	cli := models.CLI{
		Include: []string{"feat", "refactor"},
		Exclude: []string{"fix"},
	}
	patch(conf, cli)
	assert.Equal(t, conf.Changelog.Include, []string{"feat", "refactor"})
}
