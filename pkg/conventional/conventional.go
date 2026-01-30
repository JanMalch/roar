package conventional

import (
	"regexp"
	"strings"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/plumbing/storer"
	"github.com/janmalch/roar/util"
	"github.com/pkg/errors"
)

type ConventionalCommit struct {
	Commit *object.Commit
	// the conventional commit type extracted from the message
	Type string
	// the (optional) conventional commit scope extracted from the message
	Scope string
	// the human-readable part of the conventional commit message
	Title string
	// indicating if this commit is a breaking change
	BreakingChange bool
	// message for the breaking change, identified by "BREAKING CHANGE" in the body, or the title if the former isn't present
	BreakingChangeMessage string
	// the change level
	Change util.Change
}

// TODO: drop regex for parser?

// Conventional commit regex based on https://stackoverflow.com/a/62293234
var re = regexp.MustCompile(`(?m)^((?P<type>build|chore|ci|docs|feat|fix|perf|refactor|revert|style|test|¯\\_\(ツ\)_/¯)(\((?P<scope>\w+)\))?(?P<break>!)?(: (.*\s*)*))|(Merge (.*\s*)*)|(Initial commit$)`)
var (
	ErrNoCommits                 = errors.New("no commits found")
	ErrOnlyUnconventionalCommits = errors.New("only unconventional commits found")
)

func toChange(typ string, breaking bool) util.Change {
	if breaking {
		return util.MAJOR_CHANGE
	}
	if typ == "feat" {
		return util.MINOR_CHANGE
	}
	return util.PATCH_CHANGE
}

func Parse(c *object.Commit) *ConventionalCommit {
	if strings.EqualFold(c.Message, "Initial commit") {
		return &ConventionalCommit{
			Commit:         c,
			Type:           "",
			Scope:          "",
			Title:          c.Message,
			BreakingChange: false,
			Change:         util.MINOR_CHANGE,
		}
	}

	lines := strings.Split(c.Message, "\n")
	if len(lines) == 0 {
		return nil
	}
	// FIXME: verify that message is valid!
	matches := re.FindStringSubmatch(lines[0])
	if len(matches) == 0 {
		return nil
	}

	title := ""
	for i := len(matches) - 1; i > 5; i-- {
		title = strings.TrimSpace(matches[i])
		if title != "" {
			break
		}
	}

	typ := matches[re.SubexpIndex("type")]
	breaking := matches[re.SubexpIndex("break")] == "!"
	breakingMessage := title
	for _, line := range lines {
		if strings.HasPrefix(line, "BREAKING CHANGE:") {
			breaking = true
			breakingMessage = strings.TrimSpace(line[17:])
			break
		}
	}
	return &ConventionalCommit{
		Commit:                c,
		Type:                  typ,
		Scope:                 matches[re.SubexpIndex("scope")],
		Title:                 title,
		BreakingChange:        breaking,
		BreakingChangeMessage: breakingMessage,
		Change:                toChange(typ, breaking),
	}
}

// collects conventional commits from the given log in a map, where the key is the scope
//
// returns the maximum change in the log
func CollectSince(r *git.Repository, ltag *plumbing.Reference) (map[string][]*ConventionalCommit, util.Change, error) {
	iter, err := r.Log(&git.LogOptions{Order: git.LogOrderCommitterTime})
	if err != nil {
		return nil, util.NO_CHANGE, errors.Wrap(err, "Failed to get commit log")
	}
	hasAnyCommit := false
	change := util.NO_CHANGE
	commitsByScope := make(map[string][]*ConventionalCommit, 0)

	tagObj, err := r.TagObject(ltag.Hash())
	var ltagHash plumbing.Hash
	if err == nil {
		ltagHash = tagObj.Target // annotated
	} else {
		ltagHash = ltag.Hash() // lightweight
	}

	if err = iter.ForEach(func(c *object.Commit) error {
		if c.Hash.Equal(ltagHash) {
			return storer.ErrStop
		}
		hasAnyCommit = true
		if len(c.ParentHashes) > 1 {
			// Skip merge commits
			return nil
		}
		cc := Parse(c)
		if cc != nil {
			group, exists := commitsByScope[cc.Scope]
			if !exists {
				commitsByScope[cc.Scope] = []*ConventionalCommit{cc}
			} else {
				commitsByScope[cc.Scope] = append(group, cc)
			}
			if cc.Change > change {
				change = cc.Change
			}
		}
		return nil
	}); err != nil {
		return nil, util.NO_CHANGE, err
	}
	if !hasAnyCommit {
		return nil, util.NO_CHANGE, ErrNoCommits
	}
	if change == util.NO_CHANGE {
		return nil, util.NO_CHANGE, ErrOnlyUnconventionalCommits
	}
	return commitsByScope, change, nil
}
