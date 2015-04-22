package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os"
	"os/exec"
	"time"
)

const (
	timeFmt = "20060102150405"
)

func backupConf(confPath string) error {
	// check if the backup path exists, if not make it
	// todo combine into a single func util
	_, err := os.Stat(defaultBackupPath)
	if err != nil {
		os.MkdirAll(defaultBackupPath, 0777)
	}

	tarBackup := fmt.Sprintf(defaultBackupPath + "sensu-check-" + getTime() + ".tar")
	cmd := exec.Command("tar", "cvf", tarBackup, confPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("Error creating config backup for /etc/sensu/conf.d , verify sensu conf.d path: %s %s\n", out, err)
	}
	if err != nil {
		fmt.Printf("Error creating config backup: %s\n", err)
		return err
	}
	return nil
}

func getTime() string {
	t := time.Now().UTC().Local()
	fmt.Println(t.Format(timeFmt))
	return t.Format(timeFmt)
}
