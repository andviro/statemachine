package main

import (
	"bytes"
	"github.com/pkg/errors"
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
	Iter  bool   `yaml:"iter"`
	cName *tpl.Template
	cBody *tpl.Template
	cPath map[string]*tpl.Template
	isDir bool
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
	"Index": func() int {
		return -1
	},
	"Source": func() string {
		return "<unknown>"
	},
}

func (t *Template) Compile(src string) (err error) {
	if t.Body != "" && t.Path != "" {
		panic("both template body and template path specified")
	}
	t.Src = src
	if t.cName, err = tpl.New("$name").Funcs(funcMap).Parse(t.Name); err != nil {
		return errors.Wrap(err, "parse name")
	}
	if t.Body != "" {
		t.cBody, err = tpl.New("$body").Funcs(funcMap).Parse(t.Body)
		return errors.Wrap(err, "parse body")
	}
	path, err := filepath.Abs(src)
	if err != nil {
		return
	}
	t.cPath = make(map[string]*tpl.Template)
	startPath := filepath.Join(filepath.Dir(path), t.Path)
	info, err := os.Stat(startPath)
	if err != nil {
		return errors.Wrap(err, "get file info")
	}
	t.isDir = info.IsDir()
	err = filepath.Walk(startPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		var relPath string
		if relPath, err = filepath.Rel(startPath, path); err != nil {
			return err
		}
		t.cPath[relPath], err = tpl.New(relPath).Funcs(funcMap).ParseFiles(path)
		return errors.Wrap(err, "parse file")
	})
	return
}

func (t *Template) run(data interface{}, extraFuncs tpl.FuncMap) (err error) {
	buf := new(bytes.Buffer)
	if err = t.cName.Funcs(extraFuncs).Execute(buf, data); err != nil {
		return
	}
	name := buf.String()

	if t.cBody != nil {
		fp, err := os.Create(name)
		if err != nil {
			return err
		}
		defer fp.Close()
		return t.cBody.Funcs(extraFuncs).Execute(fp, data)
	}

	for k, v := range t.cPath {
		if !t.isDir {
			k = name
		} else {
			k = filepath.Join(name, k)
		}
		if err = os.MkdirAll(filepath.Dir(k), os.ModePerm); err != nil {
			return
		}
		if err = func() error {
			fp, err := os.Create(k)
			if err != nil {
				return err
			}
			defer fp.Close()
			return v.Funcs(extraFuncs).Execute(fp, data)
		}(); err != nil {
			return
		}
	}
	return

}

func (t *Template) Execute(machines []*Machine, srcFile string) (err error) {
	if t.cName == nil || (t.cBody == nil && len(t.cPath) == 0) {
		panic("template not compiled")
	}
	if !t.Iter {
		return t.run(machines, tpl.FuncMap{
			"Source": func() string { return filepath.Base(srcFile) },
		})
	}
	for idx, m := range machines {
		if err = t.run(m, tpl.FuncMap{
			"Source": func() string { return filepath.Base(srcFile) },
			"Index":  func() int { return idx },
		}); err != nil {
			return
		}
	}
	return
}
