package main

import (
	"fmt"
)

type GitParams struct {
	repo         string
	timeInterval int
	configPath   string
	backupPath   string
}

type toString interface {
	String() string
}

// GetParam accessors
func (g *GitParams) getRepo() string {
	return g.repo
}
func (g *GitParams) getTimeInterval() int {
	return g.timeInterval
}
func (g *GitParams) getConfigPath() string {
	return g.configPath
}
func (g *GitParams) getBackupPath() string {
	return g.backupPath
}
func (g *GitParams) String() string {
	s := fmt.Sprintf("Repository: [%s] \n"+
		"Polling Interval: [%d] \n"+
		"Configuration Path: [%s] \n"+
		"Configuration Backup Path: [%s] \n",
		g.getRepo(), g.getTimeInterval(), g.getConfigPath(), g.getBackupPath())
	return s
}
