package git_test

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/janmalch/roar/pkg/git"
)

func setupDirtyRepo(t *testing.T) *git.Repo {
	d := t.TempDir()
	f, err := os.Create(path.Join(d, "dummy.txt"))
	require.NoErrorf(t, err, "setup error: failed to create file in %s", d)
	defer func() {
		require.NoErrorf(t, f.Close(), "setup error: failed to close the dummy file in %s", d)
	}()
	_, err = f.WriteString("roar!\n")
	require.NoErrorf(t, err, "setup error: failed to write content the dummy file in %s", d)

	r := git.NewRepo(d)
	_, err = r.ExecGit("init", "-b", "main")
	require.NoErrorf(t, err, "setup error: failed to initialize a new git repository in %s", d)

	return r
}

func makeTestCommit(t *testing.T, r *git.Repo) {
	err := r.Add(".")
	require.NoError(t, err, "setup failed: error while adding files to git")
	err = r.Commit("test commit")
	require.NoError(t, err, "setup failed: error while commiting files")
}

func TestIsGitRepoFalse(t *testing.T) {
	r := setupDirtyRepo(t)
	isRepo := r.IsGitRepo()
	assert.True(t, isRepo, "expected the repository to be a git repository")
}

func TestIsCleanFalse(t *testing.T) {
	r := setupDirtyRepo(t)
	clean, err := r.IsClean()
	if assert.NoError(t, err) {
		assert.False(t, clean, "expected the repository to be dirty, but it was clean")
	}
}

func TestIsCleanTrue(t *testing.T) {
	r := setupDirtyRepo(t)
	makeTestCommit(t, r)

	clean, err := r.IsClean()
	if assert.NoError(t, err) {
		assert.True(t, clean, "expected the repository to be clean, but it was dirty")
	}
}

func TestHasCommitsTrue(t *testing.T) {
	r := setupDirtyRepo(t)
	makeTestCommit(t, r)

	hasCommits, err := r.HasCommits()
	if assert.NoError(t, err) {
		assert.True(t, hasCommits, "expected the repository to have commits")
	}
}

func TestHasCommitsFalse(t *testing.T) {
	r := setupDirtyRepo(t)

	hasCommits, err := r.HasCommits()
	if assert.NoError(t, err) {
		assert.False(t, hasCommits, "expected the repository to NOT have commits")
	}
}

func TestCurrentBranchName(t *testing.T) {
	r := setupDirtyRepo(t)
	makeTestCommit(t, r)

	branch, err := r.CurrentBranchName()
	if assert.NoError(t, err) {
		assert.Equal(t, "main", branch, "expected the current branch to be main")
	}
}

func TestNoTagsYet(t *testing.T) {
	r := setupDirtyRepo(t)
	makeTestCommit(t, r)
	v, err := r.LatestVersionTag()
	if assert.NoError(t, err) {
		assert.Equal(t, "", v, "expected the latest tag to be none")
	}
}

func TestExistingTags(t *testing.T) {
	r := setupDirtyRepo(t)
	makeTestCommit(t, r)
	assert.NoError(t, r.AddTag("v0.0.1"))
	v, err := r.LatestVersionTag()
	if assert.NoError(t, err) {
		assert.Equal(t, "v0.0.1", v, "expected the latest tag to be v0.0.1")
	}
	assert.NoError(t, r.AddTag("v0.0.2"))
	v, err = r.LatestVersionTag()
	if assert.NoError(t, err) {
		assert.Equal(t, "v0.0.2", v, "expected the latest tag to be v0.0.2")
	}
}

func TestContentfulCommitLogSinceTag(t *testing.T) {
	r := setupDirtyRepo(t)
	makeTestCommit(t, r)
	assert.NoError(t, r.AddTag("v0.0.1"))
	r.ExecGit("commit", "-m='second commit'", "--allow-empty")
	r.ExecGit("commit", "-m='third commit'", "--allow-empty")
	log, err := r.CommitLogSince("v0.0.1")
	if assert.NoError(t, err) {
		assert.Len(t, log, 2, "expected the log to contain two commits")
	}
}

func TestContentfulCommitLogSinceInitial(t *testing.T) {
	r := setupDirtyRepo(t)
	makeTestCommit(t, r)
	r.ExecGit("commit", "-m='second commit'", "--allow-empty")
	r.ExecGit("commit", "-m='third commit'", "--allow-empty")
	log, err := r.CommitLogSince("")
	if assert.NoError(t, err) {
		assert.Len(t, log, 3, "expected the log to contain three commits")
	}
}
