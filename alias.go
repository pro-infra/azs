package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
)

const configfile = "~/.config/azs.json"

type aliasConfig struct {
	Subscriptions map[string]string `json:"subscriptions"`
}

func newAliases() aliasConfig {
	c := aliasConfig{}
	p, err := homedir.Expand(configfile)
	if err != nil {
		panic(err)
	}
	b, err := os.ReadFile(p)

	if errors.Is(err, os.ErrNotExist) {
		c.Subscriptions = map[string]string{}
		return c
	}
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(b, &c); err != nil {
		panic(err)
	}
	return c
}

func (c *aliasConfig) store(s subscription) {
	c.Subscriptions[s.ID] = s.Name
	p, err := homedir.Expand(configfile)
	if err != nil {
		panic(err)
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(p, b, 0644); err != nil {
		panic(err)
	}
}

func (c *aliasConfig) get(s subscription) subscription {
	s.Name = fmt.Sprintf("z - %s", s.OrginalName)
	if name, ok := c.Subscriptions[s.ID]; ok {
		s.Name = name
	}
	return s
}
