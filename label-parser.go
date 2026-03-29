package docker

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type LabelParser struct {
	groupSelector *Selector
	typeSelector  *Selector
}

func NewLabelParser(prefix string, separator string, useGroups bool) (*LabelParser, error) {
	prefixEscaped := regexp.QuoteMeta(prefix)
	separatorEscaped := regexp.QuoteMeta(separator)

	const groupGroupName = "group"
	const typeGroupName = "type"

	var groupPattern string
	if useGroups {
		groupPattern = fmt.Sprintf(`%s(?P<%s>\S+)`, separatorEscaped, groupGroupName)
	}
	typePattern := fmt.Sprintf(`%s(?P<%s>\S+)`, separatorEscaped, typeGroupName)

	pattern, err := regexp.Compile(
		fmt.Sprintf(`^%s%s%s$`, prefixEscaped, groupPattern, typePattern),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to compile label parsing regex: %w", err)
	}

	parser := &LabelParser{
		typeSelector: &Selector{
			matcher:       *pattern,
			selectorGroup: typeGroupName,
		},
	}

	if useGroups {
		parser.groupSelector = &Selector{
			matcher:       *pattern,
			selectorGroup: groupGroupName,
		}
	}

	return parser, nil
}

func (r *LabelParser) parseLabel(label string) (string, string, error) {

	if r.typeSelector == nil {
		return "", "", errors.New("type selector is required")
	}

	typeString, err := r.typeSelector.Select(label)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse type from label %q: %w", label, err)
	}

	if r.groupSelector != nil {
		groupString, err := r.groupSelector.Select(label)
		if err != nil {
			return "", "", fmt.Errorf("failed to parse group from label %q: %w", label, err)
		}
		return strings.ToLower(typeString), strings.ToLower(groupString), nil
	}

	return strings.ToLower(typeString), "", nil
}

func (r *LabelParser) ParseLabelGroups(labels map[string]string) map[string]map[string]string {

	temp := make(map[string]map[string]string)

	for label, value := range labels {
		labelType, labelGroup, err := r.parseLabel(label)
		if err != nil {
			//Logger.Debugf("Skipping label %q: %v", label, err)
			continue
		}

		if _, ok := temp[labelGroup]; !ok {
			temp[labelGroup] = make(map[string]string)
		}

		temp[labelGroup][labelType] = strings.TrimSpace(value)
	}

	return temp
}
