package main

import (
	"fmt"
	"sort"
)

type M map[string]interface{}

type Machine struct {
	Name   string  `yaml:"name"`
	Title  string  `yaml:"title"`
	Meta   M       `yaml:"meta"`
	States []State `yaml:"states"`
}

type State struct {
	Name   string  `yaml:"name"`
	Title  string  `yaml:"title,omitempty"`
	Meta   M       `yaml:"meta"`
	Events []Event `yaml:"events"`
}

type Event struct {
	Name    string   `yaml:"name"`
	Next    string   `yaml:"next"`
	Meta    M        `yaml:"meta"`
	Actions []string `yaml:"actions,omitempty"`
}

func (m *Machine) UnknownStates() (res []string) {
	knownStates := make(map[string]bool)
	unknownStates := make(map[string]bool)
	for _, s := range m.States {
		knownStates[s.Name] = true
		if unknownStates[s.Name] {
			delete(unknownStates, s.Name)
		}
		for _, a := range s.Events {
			if a.Next == "" {
				continue
			}
			if !knownStates[a.Next] {
				unknownStates[a.Next] = true
			}
		}
	}
	if len(unknownStates) == 0 {
		return
	}
	fmt.Print(unknownStates)
	res = make([]string, 0, len(unknownStates))
	for k := range unknownStates {
		res = append(res, k)
	}
	sort.Strings(res)
	return
}
