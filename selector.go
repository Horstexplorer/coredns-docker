package docker

import (
	"fmt"
	"regexp"
)

type Selector struct {
	matcher       regexp.Regexp
	selectorGroup string
}

func (r *Selector) Matches(value string) bool {
	return r.matcher.MatchString(value)
}

func (r *Selector) Select(value string) (string, error) {
	if !r.Matches(value) {
		return "", fmt.Errorf("value %q does not match selector pattern", value)
	}
	selectorPos := r.matcher.SubexpIndex(r.selectorGroup)
	if selectorPos <= 0 {
		return "", fmt.Errorf("invalid selector group name %q", r.selectorGroup)
	}
	matches := r.matcher.FindStringSubmatch(value)
	if selectorPos >= len(matches) {
		return "", fmt.Errorf("selector group %q not found in matches", r.selectorGroup)
	}
	return matches[selectorPos], nil
}
