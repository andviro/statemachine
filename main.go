package main

import (
	"gopkg.in/yaml.v2"

	//"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
)

var templateOverride = flag.String("t", "", "Override template in YAML")

func init() {
	flag.Parse()
}

func main() {
	if len(os.Args) < 2 {
		os.Exit(1)
	}
	data, err := ioutil.ReadFile(flag.Arg(0))
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

	for _, m := range machines {
		var tplName string
		if *templateOverride != "" {
			tplName = *templateOverride
		} else {
			tplName = m.Template
		}
		data, err := ioutil.ReadFile(tplName)
		if err != nil {
			log.Fatalf("template load error: %+v", err)
		}
		var templates []*Template
		if err = yaml.Unmarshal(data, &templates); err != nil {
			log.Fatalf("template parse error: %+v", err)
		}
		for _, t := range templates {
			if err := t.Compile(tplName); err != nil {
				log.Printf("template compile error: %v", err)
				continue
			}
			if err := t.Execute(m); err != nil {
				log.Printf("template execute error: %v", err)
				continue
			}
		}
	}
}
