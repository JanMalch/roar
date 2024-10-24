package steps

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/janmalch/roar/pkg/git"
	"github.com/pkg/errors"
)

var ErrReleaseAsNotAfterLatest = errors.New("release-as version is not after latest")

func DetermineLatest(r *git.Repo) (string, *semver.Version, error) {
	tag, err := r.LatestVersionTag()
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to determine latest tag")
	}
	if tag == "" {
		return "", nil, nil
	} else {
		latest, err := semver.NewVersion(tag[1:])
		if err != nil {
			return tag, nil, fmt.Errorf("failed to parse latest tag %s as semver version", tag[1:])
		} else {
			return tag, latest, nil
		}
	}
}
