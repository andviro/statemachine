package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	tpl "text/template"
)

var idCleanupRe = regexp.MustCompile("[^a-zA-Z0-9]+")

type Template struct {
	Src   string `yaml:"-"`
	Name  string `yaml:"name"`
	Body  string `yaml:"body"`
	Path  string `yaml:"path"`
	cName *tpl.Template
	cBody *tpl.Template
	cPath map[string]*tpl.Template
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

var funcMap = tpl.FuncMap{
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

func (t *Template) Compile(src string) (err error) {
	if t.Body != "" && t.Path != "" {
		panic("both template body and template path specified")
	}
	t.Src = src
	if t.cName, err = tpl.New("").Funcs(funcMap).Parse(t.Name); err != nil {
		return
	}
	if t.Body != "" {
		t.cBody, err = tpl.New("").Funcs(funcMap).Parse(t.Body)
		return
	}
	path, err := filepath.Abs(src)
	if err != nil {
		return
	}
	base := filepath.Join(filepath.Dir(path), t.Src)
	err = filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		fmt.Println(path, info, err)
		return err
	})
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
