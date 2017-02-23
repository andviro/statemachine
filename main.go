package main

import (
	"gopkg.in/yaml.v2"

	//"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
)

var template = flag.String("t", "", "Override template in YAML")

func init() {
	flag.Parse()
}

func main() {
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

	for i, t := range templates {
		if err := t.Compile(*template); err != nil {
			log.Printf("template %d compile error: %v", i, err)
			continue
		}
		if err := t.Execute(machines, srcFile); err != nil {
			log.Printf("template %d execute error: %v", i, err)
			continue
		}
	}
}
