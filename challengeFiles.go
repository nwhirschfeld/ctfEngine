package main

import (
	"crypto/sha256"
	"fmt"
	"os"
)

type challengeFile struct {
	Filename string
	Size     int64
	Location string
}

func readChallengeFile(path string) (challengeFile, error) {
	var chaFi challengeFile
	chaFi.Location = path
	fi, err := os.Stat(path)
	if err != nil {
		return challengeFile{}, err
	}
	chaFi.Size = fi.Size()
	chaFi.Filename = fi.Name()

	return chaFi, nil
}

func readChallengeFiles(path string) map[string]challengeFile {
	challengeFiles := make(map[string]challengeFile)

	challengeFilesOS, err := os.ReadDir(path)

	if err == nil {
		/* challenge files directory exists */
		for _, challengeFileOS := range challengeFilesOS {
			if !challengeFileOS.IsDir() {
				challengeFile, err := readChallengeFile(fmt.Sprintf("%s/%s", path, challengeFileOS.Name()))
				if err == nil {
					hash := sha256.New()
					hash.Write([]byte(challengeFileOS.Name()))
					challengeFiles[fmt.Sprintf("%x", hash.Sum(nil))] = challengeFile
				}
			}
		}

	}
	return challengeFiles
}
