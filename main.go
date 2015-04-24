package main

// Example usage:
// sensu-git \
// -t 60 \
// -l debug \
// -r https://github.com/nerdalert/plugin-watch.git \
// -c etc/sensu/conf.d \
// -b etc/sensu/conf.d.backups

import (
	log "github.com/nerdalert/sensu-git/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/nerdalert/sensu-git/Godeps/_workspace/src/github.com/jessevdk/go-flags"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

const (
	defaultInterval    = 90
	defaultIntervalMin = 20
	sensuSrvrService   = "sensu-server"
	tempDir            = "tmp/plugin-watch"
	defaultSensuPath   = "/etc/sensu/conf.d"
	defaultBackupPath  = "/etc/sensu/conf.d.backups"
)

// only daemon mode supported atm. The -d flag is ignored
var opts struct {
	GitRepo          string `short:"r" long:"repo" description:"(required) target repository url - example format: https://github.com/nerdalert/plugin-watch.git"`
	TimeInterval     int    `short:"t" long:"time" description:"(requiredl) time in seconds between Git repository update checks."`
	CheckConfigPath  string `short:"c" long:"config-path" description:"(recommended: default: [/etc/sensu/conf.d/]) path to the sensu 'check' config files."`
	ConfigBackupPath string `short:"b" long:"backup-path" description:"(recommended: default: [/etc/sensu/conf.d.backups/]) path to the backup sensu 'check' config files."`
	Daemon           bool   `short:"d" long:"daemon" description:"(optional:default [true]) run as a daemon. Alternatively could be run via a cron job."`
	LogLevel         string `short:"l" long:"loglevel" description:"(optional:default [info]) set the logging level. Options are [debug, info, warn, error]."`
	Help             bool   `short:"h" long:"help" description:"show app help."`
}

func init() {
	runtime.GOMAXPROCS(1)
	ch := make(chan os.Signal, 1)
	go sigHandler(ch)
}

func sigHandler(ch chan os.Signal) {
	signal.Notify(ch, os.Interrupt, syscall.SIGKILL, syscall.SIGTERM)
	go func() {
		for _ = range ch {
			os.Exit(0)
		}
	}()
}

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}
	if opts.Help {
		showUsage()
		os.Exit(1)
	}
	if opts.GitRepo == "" {
		showUsage()
		log.Fatal("Required repo name is missing")
		os.Exit(1)
	}
	if opts.TimeInterval < defaultIntervalMin {
		showUsage()
		log.Fatal("The minimum polling interval is 30 seconds.")
		os.Exit(1)
	}
	// Set logrus logging level, default is Info
	switch opts.LogLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.InfoLevel)
		log.Debug("Logging level is set to : ", log.GetLevel())
	}
	g := NewGitParam()
	log.Debugf("connecting to [ %s ] with the following paramters: \n%s ", g.getRepo(), g)
	g.run()
}

func NewGitParam() *GitParams {
	var confDir string
	confDir = opts.CheckConfigPath
	if opts.CheckConfigPath == "" {
		confDir = defaultSensuPath
	}
	// verify if the sensu/conf.d dir exists, fail if it doesn't
	src, err := os.Stat(confDir)
	if err != nil {
		log.Fatalf("[%v] does not exist, specify the sensu/conf.d path with a '-p' flag", confDir)
		os.Exit(1)
	}
	if !src.IsDir() {
		log.Printf("[%v] is not a directory or does not exist, specify the sensu/conf.d path with a '-p' flag", confDir)
		os.Exit(1)
	}

	var backupDir string
	backupDir = opts.ConfigBackupPath
	if opts.ConfigBackupPath == "" {
		backupDir = defaultBackupPath
	}
	// check if the sensu conf.d.backup path exists, create it if it doesn't
	_, err = os.Stat(backupDir)
	if err != nil {
		log.Debug("The sensu config.d directory was not found in the path [%s]", backupDir)
		err := os.MkdirAll(backupDir, 0777)
		if err != nil {
			log.Debug("Unable to create the sensu backup path [%s] : %s", backupDir, err)
		}
	}

	var timeInterval int
	timeInterval = opts.TimeInterval
	if opts.TimeInterval == 0 {
		timeInterval = defaultInterval
		log.Debug("Polling interval not specified, setting it to 90 seconds")
	}
	return &GitParams{
		repo:         opts.GitRepo,
		timeInterval: timeInterval,
		configPath:   confDir,
		backupPath:   backupDir,
	}
}

func showUsage() {
	var usage string
	usage = `
Usage:
  main

Application Options:
    -r, --repo=         (required) target repository url - example format: https://github.com/nerdalert/plugin-watch.git
    -t, --time=         (requiredl) time in seconds between Git repository update checks.
    -c, --config-path=  (recommended: default: [/etc/sensu/conf.d/]) path to config files.
    -b, --backup-path=  (recommended: default: [/etc/sensu/conf.d.backups/]) path to the backup sensu 'check' config files.
    -s, --server=       (optional: default: [/etc/sensu/conf.d/]) path to config files.
    -d, --daemon=       (optional:default [true]) run as a daemon. Alternatively could be run via a cron job.
    -l, --loglevel=     (optional:default [info]) set the logging level. Default is 'info'. options are [debug, info, warn, error].
    -h, --help    show app help.

Example daemon mode processing flows every 2 minutes:
	sensu-git -r github.com/plugin-watch -t 120 -l debug -r https://github.com/nerdalert/plugin-watch.git

Example run-once export:
    TODO:

Help Options:
  -h, --help    Show this help message
  `
	log.Print(usage)
}
