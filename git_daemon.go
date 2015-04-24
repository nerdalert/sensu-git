package main

import (
	"bytes"
	"fmt"
	log "github.com/nerdalert/sensu-git/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Simple FSM as follows:
// 1) Prior to the daemon being called, CLI arguments are bound
// to GitParams and whatever is null is initialized to a default.

// 2) In the run() method prior to the timer ticking the sensu host
// clones the repo if it does not already exist via the process()
// method. The repo is cloned to a scratch directory in /tmp.
//
// 3) If the repo doesn't exist, it is cloned and updated to the
//  temp dir (currently defined as a constant). If the repo already
// exists, it is updated (git pull).
//
// 4) At startup, there is a one-time copy from the cloned repo to
// the sensu configuration directory to the user specified sensu
// configuration directory or if left blank the default path
// in /etc/sensu/conf.d
//
// 5) Subsequent changes are only triggered by modifications to
// the upstream git repository. It is currently hardcoded to the
// master branch but there are methods to read the current branch
// if desired.
//
// 6) When a change is triggered from the upstream repo, the local
// repo is synced to upstream. Those files are then copied to the
// sensu conf.d (or wherever specified by the runtime argument).
// Anytime the conf files are copied into the production conf dir
// a backup is gzipped with a timestamp prior to any potential
// overwriting of the conf dir.
//
// 7) Now that the new foo.json service check config is on the
// sensu server, the sensu-server service needs to be reloaded.
// In current form, the sensu-server proc is bounced via
// supervisord using the supervisorctl client command.

// 8) There is a file stat watcher on the conf.d directory that
// that is triggered via a syscall file stat event. The result is
// the service check gets distributed via rabbitmq to all endpoints
// subscribed to the queue the check is subscribed in the config.

// TODO: support non-supervisord.
// TODO: add some config check of check/conf.json fields.
// TODO: (cont) since the config json data models are are fairly
// TODO: (cont) different, specific checks will reduce flexibility.

// run the the simple fsm
func (g *GitParams) run() {
	// set a timer
	timer := time.NewTicker(time.Duration(g.getTimeInterval()) * time.Second)
	// create the temp dir where the repo will clone to
	err := g.setupTempDir()
	if err != nil {
		log.Fatalf("unable to create a temp directory in [%s]"+
			" verify the uid has proper permissions", tempDir)
	}
	// backup existing config
	err = g.backupConf()
	if err != nil {
		log.Errorf("sensu config was not backed up from [%s] error: ", g.getConfigPath(), err)
	}
	// process git operations
	g.process()
	// copy repo conf files to sensu/conf.d directory
	err = g.copySensuChecks()
	if err != nil {
		log.Errorf("Unable to copy check files from [%s] "+
			" to the directory [%s]", tempDir, g.getConfigPath())
	}
	err = g.restartSensuProc()
	if err != nil {
		log.Fatalf("Unable to restart the sensu-server service [%s]"+
			" verify the uid has proper permissions: %s", err)
	}
	for {
		// fork a go routine to watch for config updates
		doneChan := make(chan bool)
		go func(doneChan chan bool) {
			defer func() {
				doneChan <- true
			}()
			err := watchDir(tempDir)
			//		err := watchDir(defaultSensuPath)
			if err != nil {
				fmt.Println(err)
			}
			log.Debug("A file has changed")
		}(doneChan)
		// wait for timer to expire or a file watch event
		for b := true; b; {
			select {
			case <-doneChan:
				log.Debugf("File change event occoured at [%s] ", time.Now().Format("16:04:05"))
				continue
			case <-timer.C:
				log.Debug("Interval time expired")
				// process git operations
				g.process()
				b = false
			}
		}
		log.Debugf("checking repo for updates %s ", time.Now().Format("16:04:05"))
	}
}

func (g *GitParams) process() {
	git := newGit(tempDir)
	gitExists := fmt.Sprintf("%s/.git", tempDir)
	if _, err := os.Stat(gitExists); err != nil {
		c := git.clone(g.getRepo())
		err := c.Run()
		if err != nil {
			log.Fatalf("Error cloning git repo [%s] aborting:  %s\n", g.getRepo(), err)
		}
	}
	cmdOutput := &bytes.Buffer{}
	c := git.update()
	c.Stdout = cmdOutput
	err := c.Run()
	if err != nil {
		log.Fatalf("Error updating repo: %s\n", err)
	}
	debugGitCmd(cmdOutput.Bytes())
	if !strings.Contains(string(cmdOutput.Bytes()), "Already up-to-date") {
		err := g.backupConf()
		if err != nil {
			log.Fatalf("Error backing up the sensu config: %s\n", err)
		}
		g.copySensuChecks()

		if err != nil {
			log.Errorf("Unable to copy check files from [%s] to the directory [%s]", tempDir, g.getConfigPath())
		}

	}
}

func (g *GitParams) restartSensuProc() error {
	output, err := exec.Command("systemctl", "restart", sensuSrvrService).CombinedOutput()
	if err != nil {
		log.Debugf("restart failed: [%v]", output)
		return err
	}
	log.Debug("sensu-server restarted succesfully")
	return nil
}
