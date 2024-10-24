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

	err := r.Add(".")
	require.NoError(t, err, "setup failed: error while adding files to git")
	err = r.Commit("test commit")
	require.NoError(t, err, "setup failed: error while commiting files")

	clean, err := r.IsClean()
	if assert.NoError(t, err) {
		assert.True(t, clean, "expected the repository to be clean, but it was dirty")
	}
}

func TestGetOriginNone(t *testing.T) {
	r := setupDirtyRepo(t)

	origin, err := r.GetOrigin()
	if assert.NoError(t, err) {
		assert.Equal(t, "", origin, "expected the repository to have no origin")
	}
}
