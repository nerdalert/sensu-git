package main

import (
	"bytes"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
	"time"
)

// run function is a timer for processing git repo interval
func (g *GitParams) run() {
	// set a timer
	timer := time.NewTicker(time.Duration(g.getTimeInterval()) * time.Second)
	for {
		g.process()
		// wait for timer
		<-timer.C
		// log current time for debugging
		log.Debug("checking repo for updates %s ", time.Now().Format("16:04:05"))
	}
}

func (g *GitParams) process() {

	_, err := os.Stat(tempDir)
	if err != nil {
		os.MkdirAll(tempDir, 0777)
	}
	git := NewGit(tempDir)
	// check if the source dir exists
	src, err := os.Stat(git.Dir)
	if err != nil {
		log.Fatalf("[%v] does not exist, specify the sensu/conf.d path with a '-p' flag", git.Dir)
	}
	if !src.IsDir() {
		log.Printf("[%v] is not a directory or does not exist, specify the sensu/conf.d path with a '-p' flag", git.Dir)
		os.Exit(1)
	}
	gitExists := fmt.Sprintf("%s/.git", tempDir)
	if _, err := os.Stat(gitExists); err != nil {

		c := git.Clone(g.getRepo())
		err := c.Run()
		if err != nil {
			log.Fatalf("Error updating repo, aborting:  %s\n", err)
		}
	}
	cmdOutput := &bytes.Buffer{}
	c := git.Update()
	c.Stdout = cmdOutput
	err = c.Run()
	if err != nil {
		log.Fatalf("Error updating repo, aborting:  %s\n", err)
	}
	printOutput(cmdOutput.Bytes())
	if strings.Contains(string(cmdOutput.Bytes()), "Already up-to-date") {
		g.copyChecks()
	}
	backupConf(g.getConfigPath())
	err = g.copyChecks()
	if err != nil {
		log.Errorf("Unable to copy check files from [%s] to the directory [%s]", tempDir, g.configPath)
	}
}

func printOutput(outs []byte) {
	if len(outs) > 0 {
		fmt.Printf("==> Output: %s\n", string(outs))
	}
}

// copy check
func (g *GitParams) copyChecks() error {
	checkFiles := fmt.Sprintf("%s/*.json", tempDir)
	log.Infof("Stringer ---> %s", checkFiles)
	exec.Command("cp ", checkFiles, defaultSensuPath).Run()

	return nil
}
