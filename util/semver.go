package util

import (
	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
)

var (
	none                  semver.Version
	initial               semver.Version
	ErrNextNotAfterLatest = errors.New("next version is not after latest")
)

func init() {
	v, err := semver.NewVersion("0.1.0")
	if err != nil {
		panic(err)
	}
	initial = *v
}

type Change int

const (
	NO_CHANGE Change = iota
	PATCH_CHANGE
	MINOR_CHANGE
	MAJOR_CHANGE
)

func Bump(latest *semver.Version, releaseAs *semver.Version, change Change) (semver.Version, error) {
	if latest == nil {
		if releaseAs != nil {
			return *releaseAs, nil
		} else {
			return initial, nil
		}
	}

	var next semver.Version
	switch change {
	case NO_CHANGE:
		next = *latest
	case PATCH_CHANGE:
		next = latest.IncPatch()
	case MINOR_CHANGE:
		next = latest.IncMinor()
	case MAJOR_CHANGE:
		if latest.Major() == 0 {
			next = latest.IncMinor()
		} else {
			next = latest.IncMajor()
		}
	}

	if releaseAs != nil {
		if releaseAs.GreaterThan(&next) {
			return *releaseAs, nil
		} else {
			return none, ErrNextNotAfterLatest
		}
	}
	if !next.GreaterThan(latest) {
		return none, ErrNextNotAfterLatest
	}
	return next, nil
}
