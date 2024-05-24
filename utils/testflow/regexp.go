package testflow

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// var regexps = cmap.New[*regexp.Regexp]()

// func GetRegexp(pattern string) *regexp.Regexp {
// 	v, ok := regexps.Get(pattern)
// 	if ok && v != nil {
// 		return v
// 	}
// 	return regexps.Upsert(pattern, nil, func(exist bool, valueInMap *regexp.Regexp, _ *regexp.Regexp) *regexp.Regexp {
// 		if exist {
// 			return valueInMap
// 		}
// 		return regexp.MustCompile(pattern)
// 	})
// }

func Match(pattern, text string) ([][]string, error) {
	subs := regexp.MustCompile(pattern).FindAllStringSubmatch(text, -1)
	if len(subs) <= 0 {
		return nil, errors.Errorf("MatchFailed: %s", pattern)
	}
	c := make([][]string, len(subs))
	for i, x := range subs {
		c[i] = make([]string, len(x))
		for j, y := range x {
			c[i][j] = strings.Clone(y)
		}
	}
	return c, nil
}

func MatchOne(pattern, text string) ([]string, error) {
	subs, err := Match(pattern, text)
	if err != nil {
		return nil, err
	}
	if len(subs) != 1 {
		return nil, errors.Errorf("MatchOneFailed(%s): matched %d subs", pattern, len(subs))
	}
	return subs[0], nil
}

func MatchOneThen(pattern, text string, subIdx int) (string, error) {
	subs, err := MatchOne(pattern, text)
	if err != nil {
		return "", errors.Errorf("MatchOneThenFailed(%d): %s", subIdx, pattern)
	}
	if len(subs) < subIdx+1 {
		return "", errors.Errorf("MatchOneThenFailed(%d): %s", subIdx, pattern)
	}
	return subs[subIdx], nil
}

func NamedMatchOne(pattern, text string) (map[string]string, error) {
	re := regexp.MustCompile(pattern)

	subs := re.FindAllStringSubmatch(text, -1)
	if len(subs) != 1 {
		return nil, errors.Errorf("NamedMatchOneFailed(%s): matched %d subs", pattern, len(subs))
	}
	sub := subs[0]

	m := map[string]string{}
	for i, name := range re.SubexpNames() {
		if m[name] == "" {
			m[name] = sub[i]
		}
	}
	return m, nil
}
