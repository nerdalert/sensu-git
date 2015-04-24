package main

import (
	"fmt"
	log "github.com/nerdalert/sensu-git/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"os/exec"
	"time"
)

const (
	timeFmt = "20060102150405"
)

func (g *GitParams) backupConf() error {

	//	tar -cvzf tarballname.tar.gz itemtocompress
	tarBackup := fmt.Sprintf(g.getBackupPath() + "/" + "sensu-check-" + getTime() + ".tar.gz")
	cmd := exec.Command("tar", "cvzf", tarBackup, g.getConfigPath())
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("Error creating config backup for [%s] , verify sensu conf.d path: %s %s\n", g.getBackupPath(), out, err)
	}
	if err != nil {
		fmt.Printf("Error creating config backup: [%s]\n", err)
		return err
	}
	return nil
}

func getTime() string {
	t := time.Now().UTC().Local()
	fmt.Println(t.Format(timeFmt))
	return t.Format(timeFmt)
}
