package tag

import (
	"errors"
	"strings"
)

type Tag struct {
	Namespace string
	Separator string
	Label     string
}

func (t Tag) String() string {
	if t.Namespace == "" {
		return t.Label
	}
	return t.Namespace + t.Separator + t.Label
}

type parseConfig struct {
	Separator string
}

type parseOpt func(*parseConfig)

func WithSaparator(s string) parseOpt {
	return func(pc *parseConfig) { pc.Separator = s }
}

func ParseTag(tag string, opts ...parseOpt) (Tag, error) {
	conf := parseConfig{
		Separator: ":",
	}
	for _, fn := range opts {
		fn(&conf)
	}

	namespace, label, ok := strings.Cut(tag, conf.Separator)
	if !ok {
		return Tag{
			Label: tag,
		}, errors.New("tag doesn't contain any separator")
	}
	return Tag{Namespace: namespace, Separator: conf.Separator, Label: label}, nil
}
