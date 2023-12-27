package main

import (
	"fmt"
	"github.com/russross/blackfriday"
	"gopkg.in/yaml.v3"
	"html/template"
	"os"
)

type configuration struct {
	Title             string `yaml:"title"`
	Contact           string `yaml:"contact"`
	SessionTimeout    int    `yaml:"sessionTimeout"`
	CoolDown          int    `yaml:"submitCoolDown"`
	ServiceHost       string `yaml:"serviceHost"`
	RegistrationToken bool   `yaml:"registrationToken"`
	IndexPage         template.HTML
}

func readConfiguration(filePath string) (configuration, error) {
	f, err := os.ReadFile(fmt.Sprintf("%s/ctf.yml", filePath))
	if err != nil {
		return configuration{}, err
	}
	var conf configuration
	if err = yaml.Unmarshal(f, &conf); err != nil {
		return configuration{}, err
	}
	indexPage, err := os.ReadFile(fmt.Sprintf("%s/index.md", filePath))
	conf.IndexPage = ""
	if err == nil {
		conf.IndexPage = template.HTML(blackfriday.MarkdownCommon(indexPage))
	}
	return conf, nil
}
