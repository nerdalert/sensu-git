package main

import (
	"fmt"
	"github.com/coreos/fleet/log"
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
		if stat.Size() != initialStat.Size() != stat.IsDir() || stat.ModTime() != initialStat.ModTime() != stat.IsDir() {

			break
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func runWatch() {
	doneChan := make(chan bool)
	go func(doneChan chan bool) {
		defer func() {
			doneChan <- true
		}()
		err := watchDir(defaultSensuPath)
		if err != nil {
			fmt.Println(err)
		}
		log.Infof("A file has changed")
	}(doneChan)
	<-doneChan
}
