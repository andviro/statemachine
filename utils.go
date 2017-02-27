package main

import (
	"reflect"
	"regexp"
	"strings"
)

var idCleanupRe = regexp.MustCompile("[^a-zA-Z0-9]+")

func parts(s string) (res []string) {
	pts := idCleanupRe.Split(s, -1)
	for _, p := range pts {
		p = strings.TrimSpace(p)
		if len(p) == 0 {
			continue
		}
		res = append(res, p)
	}
	return
}

var funcMap = map[string]interface{}{
	"goId": func(s string) (res string) {
		for _, part := range parts(s) {
			res += strings.Title(part)
		}
		return
	},
	"pyId": func(s string) (res string) {
		return strings.Join(parts(s), "_")
	},
	"last": func(x int, a interface{}) bool {
		return x == reflect.ValueOf(a).Len()-1
	},
	"_idx": func() interface{} {
		return -1
	},
	"_srcFile": func() string {
		return "<unknown>"
	},
	"_all": func() interface{} {
		return nil
	},
}
