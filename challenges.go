package main

import (
	"crypto/sha256"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type challengeHint struct {
	Text string `yaml:"description"`
	Cost int    `yaml:"cost"`
	UID  string
}

type challengeService struct {
	Port int `yaml:"port"`
}

type challenge struct {
	Title    string `yaml:"name"`
	Text     string `yaml:"description"`
	Points   int    `yaml:"value"`
	Flag     string `yaml:"flag"`
	Category string `yaml:"category"`
	Files    map[string]challengeFile
	Hints    []challengeHint  `yaml:"hints"`
	Service  challengeService `yaml:"service"`
}

func readChallenges(path string) (map[string]challenge, error) {
	challenges := make(map[string]challenge)
	challengePath := fmt.Sprintf("%s/challenges/", path)

	items, _ := os.ReadDir(challengePath)
	for _, item := range items {
		if item.IsDir() {
			cha, err := readChallenge(filepath.Join(challengePath, item.Name()))
			if err == nil {
				challenges[item.Name()] = cha
			}
		}
	}

	return challenges, nil
}

func readChallenge(path string) (challenge, error) {
	/* read and parse challenge yml file */
	filePath := fmt.Sprintf("%s/challenge.yml", path)
	f, err := os.ReadFile(filePath)
	if err != nil {
		return challenge{}, err
	}
	var cha challenge
	if err = yaml.Unmarshal(f, &cha); err != nil {
		return challenge{}, err
	}

	// add uid to hints
	for i, hint := range cha.Hints {
		hash := sha256.New()
		hash.Write([]byte(fmt.Sprintf("%x", hint)))
		cha.Hints[i].UID = fmt.Sprintf("DINO%x", hash.Sum(nil))
	}

	cha.Files = readChallengeFiles(fmt.Sprintf("%s/files/", path))
	return cha, nil
}

func (c *challenge) print() {
	fmt.Printf("%s:%s (%d)", c.Title, c.Text, c.Points)
}
