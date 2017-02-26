package main

import (
	"bytes"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	tpl "text/template"

	"github.com/Masterminds/sprig"
)

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

func newTpl(name string) *tpl.Template {
	return tpl.New(name).Funcs(sprig.TxtFuncMap()).Funcs(funcMap)
}

func (t *Template) Compile(src string) (err error) {
	if t.Body != "" && t.Path != "" {
		return errors.New("both template body and template path specified")
	}
	t.Src = src
	if t.cName, err = newTpl(src).Parse(t.Name); err != nil {
		return
	}
	if t.Body != "" {
		t.cBody, err = newTpl(src).Parse(t.Body)
		return
	}
	path, err := filepath.Abs(src)
	if err != nil {
		return errors.Wrapf(err, "%s:1", src)
	}
	t.cPath = make(map[string]*tpl.Template)
	startPath := filepath.Join(filepath.Dir(path), t.Path)
	info, err := os.Stat(startPath)
	if err != nil {
		return errors.Wrapf(err, "%s:1", src)
	}
	t.isDir = info.IsDir()
	err = filepath.Walk(startPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return errors.Wrapf(err, "%s:1", path)
		}
		var relPath string
		if relPath, err = filepath.Rel(startPath, path); err != nil {
			return errors.Wrapf(err, "%s:1", path)
		}
		t.cPath[relPath], err = newTpl(relPath).ParseFiles(path)
		return err
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
			"_srcFile": func() string { return filepath.Base(srcFile) },
		})
	}
	for idx, m := range machines {
		if err = t.run(m, tpl.FuncMap{
			"_srcFile": func() string { return filepath.Base(srcFile) },
			"_idx":     func() int { return idx },
			"_all":     func() interface{} { return machines },
		}); err != nil {
			return
		}
	}
	return
}
