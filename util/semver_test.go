package util_test

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/janmalch/roar/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustSemver(t *testing.T, v string) *semver.Version {
	sv, err := semver.NewVersion(v)
	require.NoError(t, err)
	return sv
}

func TestDetermineNextVersionMinor(t *testing.T) {
	next, err := util.Bump(
		mustSemver(t, "0.1.0"),
		nil,
		util.MINOR_CHANGE)
	if assert.NoError(t, err) {
		assert.Equal(t, "0.2.0", next.String())
	}
}

func TestDetermineNextVersionMajor(t *testing.T) {
	next, err := util.Bump(
		mustSemver(t, "1.0.2"),
		nil,
		util.MAJOR_CHANGE)
	if assert.NoError(t, err) {
		assert.Equal(t, "2.0.0", next.String())
	}
}

func TestDetermineNextVersionMajorPrerelease(t *testing.T) {
	next, err := util.Bump(
		mustSemver(t, "0.1.0"),
		nil,
		util.MAJOR_CHANGE)
	if assert.NoError(t, err) {
		assert.Equal(t, "0.2.0", next.String())
	}
}

func TestDetermineNextVersionPatch(t *testing.T) {
	next, err := util.Bump(
		mustSemver(t, "0.1.0"),
		nil,
		util.PATCH_CHANGE)
	if assert.NoError(t, err) {
		assert.Equal(t, "0.1.1", next.String())
	}
}
