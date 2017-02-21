package main

import (
	"bytes"
	"os"
	"reflect"
	"regexp"
	"strings"
	"text/template"
)

var idCleanupRe = regexp.MustCompile("[^a-zA-Z0-9]+")

type Template struct {
	Name  string `yaml:"name"`
	Body  string `yaml:"body"`
	cName *template.Template
	cBody *template.Template
}

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

var funcMap = template.FuncMap{
	"Id": func(s string) (res string) {
		for _, part := range parts(s) {
			res += strings.Title(part)
		}
		return
	},
	"PyId": func(s string) (res string) {
		return strings.Join(parts(s), "_")
	},
	"Last": func(x int, a interface{}) bool {
		return x == reflect.ValueOf(a).Len()-1
	},
}

func (t *Template) Compile() (err error) {
	if t.cName, err = template.New("").Funcs(funcMap).Parse(t.Name); err != nil {
		return
	}
	t.cBody, err = template.New("").Funcs(funcMap).Parse(t.Body)
	return
}

func (t *Template) Execute(data interface{}) (err error) {
	if t.cName == nil || t.cBody == nil {
		panic("template not compiled")
	}
	buf := new(bytes.Buffer)
	if err = t.cName.Execute(buf, data); err != nil {
		return
	}
	fn := buf.String()
	fp, err := os.Create(fn)
	if err != nil {
		return
	}
	defer fp.Close()
	return t.cBody.Execute(fp, data)
}
