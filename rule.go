package main

import (
	"regexp"
	"strings"
)

type Rule struct {
	Exact   map[string]string
	Suffix  map[string]string
	Pattern map[string]string

	pattern map[*regexp.Regexp]string
}

func (r *Rule) Compile() error {
	if r.Exact == nil {
		r.Exact = make(map[string]string)
	}
	if r.Suffix == nil {
		r.Suffix = make(map[string]string)
	}
	if r.Pattern == nil {
		r.Pattern = make(map[string]string)
	}
	if r.pattern == nil {
		r.pattern = make(map[*regexp.Regexp]string)
	}
	for pattern, target := range r.Pattern {
		if re, err := regexp.Compile(pattern); err != nil {
			return err
		} else {
			r.pattern[re] = target
		}
	}
	return nil
}

func (r *Rule) MatchExact(name string) *string {
	if v, ok := r.Exact[name]; ok {
		return &v
	}
	return nil
}

func (r *Rule) MatchSuffix(name string) *string {
	for suffix, v := range r.Suffix {
		if strings.HasSuffix(name, suffix) {
			return &v
		}
	}
	return nil
}

func (r *Rule) MatchPattern(name string) *string {
	for re, v := range r.pattern {
		if re.MatchString(name) {
			return &v
		}
	}
	return nil
}

func (r *Rule) Match(name string) *string {
	if v := r.MatchExact(name); v != nil {
		return v
	} else if v = r.MatchSuffix(name); v != nil {
		return v
	} else if v = r.MatchPattern(name); v != nil {
		return v
	}
	return nil
}
