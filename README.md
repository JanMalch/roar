# roar

_Single-purpose CLI for opioniated semantic releases._

## Install

Download the standalone binary from the latest [v0.14.0 release page](https://github.com/JanMalch/roar/releases/tag/v0.14.0) and just run it. No installation required.

You can also use the following script. Make sure you set the correct `platform`, which must be the name of a [release asset](https://github.com/JanMalch/roar/releases/latest).

```shell
#!/bin/sh
platform="roar-arm64-darwin"
url=$(curl -s https://api.github.com/repos/JanMalch/roar/releases/latest | grep "browser_download_url.*${platform}" | cut -d : -f 2,3 | tr -d '"' | xargs echo -n)
# output in same directory as this script, ignoring current working directory
curl -sS -L -o "$(dirname "$0")/roar" "$url"
chmod +x "$(dirname "$0")/roar"
# for OS X (optional), see https://superuser.com/a/28400
# xattr -d com.apple.quarantine "$(dirname "$0")/roar"
roar -v
```

Verify it works by running `roar -v` or `roar -h`.

## Usage

Getting started is as simple as just running the CLI in a git project.
Since no configuration file is present, it will generate the default and make a dry-run with it.
The default configuration is for OpenAPI files indented with 2 spaces... so you'll likely have to make changes to the newly created `release.toml`.

A full configuration might look like this:

```toml
branch = "main" # optional: verify you are on the correct branch. Also supports regex by starting with a ^ like branch = "^release/"

[[update]] # you can define as many updates, as you like
file = "openapi.yml" # required: the file to make changes to
find = "  version: " # required: a string to detect the line to update. If it starts with a ^ it is interpreted as regex in GoLang syntax. Otherwise it is used as line prefix.
replace = "  version: {{version}}" # required: the content to replace the detected line with. Must contain the "{{version}}" placeholder.
# Other placeholders:
# {{epoch}} - timestamp of the release as UNIX epoch in milliseconds

[changelog]
include = ["feat", "fix"] # optional: which conventional commit types to include in the generated changelog. Also defines the order in the changelog.
url_commit = "https://github.com/my-org/my-repo/commit/{{hash}}"
url_browse_at_tag = "https://github.com/my-org/my-repo/tree/v{{version}}"
url_compare_tags = "https://github.com/my-org/my-repo/compare/v{{previous}}...v{{version}}"
url_commits_for_tag = "https://github.com/my-org/my-repo/commits/v{{version}}"
```

Running `roar` will look like this:

```
√ current branch 'main' matches 'main' pattern
√ get fetch --tags
i determined latest version to be v0.2.0

√ determined next version to be v0.2.1
√ updated version in openapi.yml
√ updated CHANGELOG.md
√ commited as chore(release): release version v0.2.1
√ tagged as v0.2.1

i please verify the applied changes and finalize the release by running
        git push --follow-tags
```

Run `roar -h` for extensive information.

## The name

- Release OpenApi contracts Rapidly
- Release Of yet Another Release
- Release Often and Rapidly
- Release Orchestrated And Reliable
- Release On Automated Rules
- Rapidly Organize And Release
