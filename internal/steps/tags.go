package steps

import (
	"github.com/Masterminds/semver"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/pkg/errors"
)

var ErrReleaseAsNotAfterLatest = errors.New("release-as version is not after latest")

// Determines the greatest semver tag.
func DetermineLatest(r *git.Repository) (*plumbing.Reference, *semver.Version, error) {
	iter, err := r.Tags()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to determine latest tag")
	}
	var res *plumbing.Reference
	var latest *semver.Version
	err = iter.ForEach(func(ref *plumbing.Reference) error {
		name := ref.Name().Short()
		if name[0] != 'v' {
			return nil
		}
		psmv, err := semver.NewVersion(name[1:])
		if err != nil {
			return err
		}
		if latest == nil || psmv.GreaterThan(latest) {
			res = ref
			latest = psmv
		}
		return nil
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to determine latest tag")
	}
	return res, latest, nil
}
