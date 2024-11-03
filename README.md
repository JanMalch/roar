# roar

_Single-purpose CLI for opioniated semantic releases._

## Install

Download the standalone binary from the releases page and just run it. No installation required.

## Usage

Getting started is as simple as just running the CLI in a git project.
Since no configuration file is present, it will generate the default and make a dry-run with it.
The default configuration is for OpenAPI files indented with 2 spaces... so you'll likely have to make changes to the newly created `release.toml`.

A full configuration might look like this:

```toml
file = "openapi.yml" # required: the file to make changes to
find = "  version: " # required: a string to detect the line to update. If it starts with a ^ it is interpreted as regex in GoLang syntax. Otherwise it is used as line prefix.
replace = "  version: {{version}}" # required: the content to replace the detected line with. Must contain the "{{version}}" placeholder.

include = ["feat", "fix", "refactor"] # optional: which conventional commit types to include in the generated changelog. Also defines the order in the changelog.

git_service = "gitlab" # optional: configure the used git service for links in the changelog. Should be one of "github", "gitlab", or "bitbucket".
git_service_url = "https://internal-git.mycompany.com" # optional: configure the URL if the CLI fails to properly infer it from the git origin.
```

Running `roar` will look like this:

```
i current branch is main
√ git pull
i determined latest version to be v0.2.0

√ determined next version to be v0.2.1
√ updated version in cmd/cli/main.go
√ updated CHANGELOG.md
√ commited as chore(release): release version v0.2.1
√ tagged as v0.2.1

i please verify the applied changes and finalize the release by running
        git push && git push --tags
```



## The name

- Release OpenApi contracts Rapidly
- Release Of yet Another Release
