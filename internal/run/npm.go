package run

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/janmalch/roar/models"
	"github.com/pkg/errors"
)

func npmVersion(cwd string, next semver.Version, cfg models.NpmConfig, dryrun bool) (string, error) {
	args := []string{"version", next.String(), "--git-tag-version=false"}
	if cfg.Workspaces {
		args = append(args, "--workspaces=true")
	}
	if cfg.IncludeWorkspaceRoot {
		args = append(args, "--include-workspace-root=true")
	}
	if !dryrun {
		cmd := exec.Command("npm", args...)
		if cwd != "" {
			cmd.Dir = cwd
		}
		_, err := cmd.Output()
		if err != nil {
			return "", errors.Wrap(err, fmt.Sprintf("failed command \"npm %s\"", strings.Join(args, " ")))
		}
	}
	return "npm " + strings.Join(args, " "), nil
}
