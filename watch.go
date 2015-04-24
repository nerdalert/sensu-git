package main

import (
	"fmt"
	log "github.com/nerdalert/sensu-git/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"os"
	"time"
)

func watchDir(fPath string) error {
	initialStat, err := os.Stat(fPath)
	if err != nil {
		return err
	}
	for {
		stat, err := os.Stat(fPath)
		if err != nil {
			return err
		}
		if stat.Size() != initialStat.Size() || stat.ModTime() != initialStat.ModTime() {

			break
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func runWatch() {
	log.Debugf("Starting watch on directory [%s]", tempDir)
	doneChan := make(chan bool)
	go func(doneChan chan bool) {
		defer func() {
			doneChan <- true
		}()
		err := watchDir(tempDir)
		if err != nil {
			fmt.Println(err)
		}
	}(doneChan)
	<-doneChan
}
