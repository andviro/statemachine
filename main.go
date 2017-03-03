package main

import (
	"gopkg.in/yaml.v2"

	"flag"
	"io/ioutil"
	logging "log"
	"os"
	"path/filepath"
	"strings"
)

var template = flag.String("t", "", "Override template in YAML")

func init() {
	flag.Parse()
}

func main() {
	log := logging.New(os.Stderr, "", 0)
	if len(os.Args) < 2 {
		os.Exit(1)
	}
	srcFile := flag.Arg(0)
	data, err := ioutil.ReadFile(srcFile)
	if err != nil {
		log.Fatalf("load error: %+v", err)
	}
	var machines []*Machine
	if err = yaml.Unmarshal(data, &machines); err != nil {
		log.Fatalf("parse error: %+v", err)
	}
	for _, m := range machines {
		u := m.UnknownStates()
		if len(u) != 0 {
			log.Fatalf("unresolved state references in machine '%s': %+v", m.Name, u)
		}
	}

	if *template == "" {
		flag.Usage()
		os.Exit(1)
	}

	data, err = ioutil.ReadFile(*template)
	if err != nil {
		log.Fatalf("templates load error: %+v", err)
	}
	var templates []*Template
	if err = yaml.Unmarshal(data, &templates); err != nil {
		log.Fatalf("templates parse error: %+v", err)
	}

	sourceBase := strings.TrimSuffix(srcFile, filepath.Ext(srcFile))
	for _, t := range templates {
		if err := t.Compile(*template); err != nil {
			log.Printf("%v", err)
			continue
		}
		if err := t.Execute(machines, sourceBase); err != nil {
			log.Printf("%v", err)
			continue
		}
	}
}
