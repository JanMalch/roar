package models

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	File          string   `toml:"file"`
	Find          string   `toml:"find"`
	Replace       string   `toml:"replace"`
	Include       []string `toml:"include"`
	GitService    string   `toml:"git_service"`
	GitServiceUrl string   `toml:"git_service_url"`
}

var defaultConf = Config{
	File:    "openapi.yml",
	Find:    "  version: ",
	Replace: "  version: {{version}}",
	Include: []string{"feat", "fix", "refactor"},
}

var headerText = `# Configuration for the roar CLI
# https://github.com/JanMalch/roar

`

func ConfigFromFile(path string) (*Config, bool, error) {
	var conf Config
	_, err := toml.DecodeFile(path, &conf)

	if err != nil && !os.IsNotExist(err) {
		return nil, false, err
	}
	if os.IsNotExist(err) {
		d, err := toml.Marshal(defaultConf)
		if err != nil {
			return nil, false, err
		}
		c := headerText + string(d)
		if err := os.WriteFile(path, []byte(c), 0644); err != nil {
			return nil, false, err
		}
		return &defaultConf, true, nil
	}

	if len(conf.Include) == 0 {
		conf.Include = defaultConf.Include
	}

	return &conf, false, nil
}
