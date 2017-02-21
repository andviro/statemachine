package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
)

type Machine struct {
	Name     string  `yaml:"name"`
	Title    string  `yaml:"title"`
	Template string  `yaml:"template"`
	States   []State `yaml:"states"`
}

func main() {

}
