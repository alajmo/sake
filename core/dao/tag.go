package dao

import (
	"strings"

	"github.com/alajmo/sake/core"
)

type Tag struct {
	Name    string
	Servers []string
}

func (t Tag) GetValue(key string, _ int) string {
	switch key {
	case "Tag", "tag":
		return t.Name
	case "Server", "server":
		return strings.Join(t.Servers, "\n")
	}

	return ""
}

func (c Config) GetTags() []string {
	tags := []string{}
	for _, server := range c.Servers {
		for _, tag := range server.Tags {
			if !core.StringInSlice(tag, tags) {
				tags = append(tags, tag)
			}
		}
	}

	return tags
}

func (c Config) GetTagAssocations(tags []string) ([]Tag, error) {
	t := []Tag{}

	for _, tag := range tags {
		servers, err := c.GetServersByTags([]string{tag})
		if err != nil {
			return []Tag{}, err
		}

		var serverNames []string
		for _, p := range servers {
			serverNames = append(serverNames, p.Name)
		}

		t = append(t, Tag{Name: tag, Servers: serverNames})
	}

	return t, nil
}
