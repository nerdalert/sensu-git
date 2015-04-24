package main

import (
	"fmt"
	log "github.com/nerdalert/sensu-git/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"io"
	"os"
	"strings"
)

func (g *GitParams) copySensuChecks() error {

	d, err := os.Open(tempDir)
	if err != nil {
		log.Fatalf("Error opening the temp directory [%s] : %v", tempDir, err)
	}
	srcFiles, err := d.Readdir(-1)
	d.Close()
	if err != nil {
		return err
	}
	for _, sfile := range srcFiles {
		if strings.Contains(sfile.Name(), "json") {
			newFile := fmt.Sprintf("%s/%s", g.getConfigPath(), sfile.Name())
			log.Printf("40 newFile: %s\n", newFile)
			g.copyFile(sfile.Name(), newFile)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *GitParams) copyFile(sFile, dstFile string) {

	srcPath := fmt.Sprintf("%s/%s", tempDir, sFile)
	log.Debugf("Copying the repo's check file [%s] to the sensu config directory [%s]\n", srcPath, dstFile)
	src, err := os.Open(srcPath)
	if err != nil {
		panic(err)
		os.Exit(1)
	}
	defer src.Close()
	dst, err := os.OpenFile(dstFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
		os.Exit(1)
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	if err != nil {
		panic(err)
		os.Exit(1)
	}
}

// copy check conf files
func (g *GitParams) setupTempDir() error {
	// setup a temporary directory for file diffs
	_, err := os.Stat(tempDir)
	if err != nil {
		err = os.MkdirAll(tempDir, 0777)
		if err != nil {
			return err
		}
	}
	return nil
}
