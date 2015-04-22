package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
)

// Git Create Obj
type GitCmd struct {
	Dir string
}

// NewGit create default Git object
func NewGit(dir string) *GitCmd {
	return &GitCmd{Dir: dir}
}

// Update is update command of repo
func (g *GitCmd) Update() (cmd *exec.Cmd) {
	args := []string{"pull"}
	cmd = gitCmd(args)
	cmd.Dir = g.Dir
	return
}

// Update is update command of repo
func (g *GitCmd) Fetch() (cmd *exec.Cmd) {
	args := []string{"fetch", "origin"}
	cmd = gitCmd(args)
	cmd.Dir = g.Dir
	return
}

// UpdateCurrent update the current branch
func (g *GitCmd) UpdateCurrent() (cmd *exec.Cmd) {
	args := []string{"pull", "origin", currentBranch(g.Dir)}
	cmd = gitCmd(args)
	cmd.Dir = g.Dir
	return
}

func (g *GitCmd) Clone(repo string) (cmd *exec.Cmd) {
	args := []string{"clone", repo, g.Dir}
	cmd = gitCmd(args)
	return
}

func (g *GitCmd) HasRemote() bool {
	args := []string{"config", "--get", "remote.origin.url"}
	cmd := gitCmd(args)
	cmd.Dir = g.Dir
	err := cmd.Run()
	return err == nil
}

func currentBranch(path string) string {
	args := []string{"rev-parse", "--abbrev-ref", "HEAD"}
	cmd := exec.Command("git", args...)
	output := new(bytes.Buffer)
	cmd.Stdout = output
	cmd.Stderr = output
	cmd.Dir = path
	err := cmd.Run()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(output.String())
}

func gitCmd(args []string) (cmd *exec.Cmd) {
	cmd = exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return
}
