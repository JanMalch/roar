package steps

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

var (
	ErrNoFindResult = errors.New("go no results with given find string")
)

func FindAndReplace(path, find, replacement string, dryrun bool) error {
	var err error
	var r *regexp.Regexp

	if strings.HasPrefix(find, "^") {
		r, err = regexp.Compile(find)
		if err != nil {
			return err
		}
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(b)))
	var updated strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if r != nil {
			if r.MatchString(line) {
				updated.WriteString(replacement)
			} else {
				updated.WriteString(line)
			}
		} else {
			if strings.HasPrefix(line, find) {
				updated.WriteString(replacement)
			} else {
				updated.WriteString(line)
			}
		}
		updated.WriteRune('\n')
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if !dryrun {
		return os.WriteFile(path, []byte(updated.String()), 0644)
	} else {
		return nil
	}
}
