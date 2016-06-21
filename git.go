package git

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/desal/cmd"
	"github.com/desal/dsutil"
)

type Status int

const (
	Unknown Status = iota
	Clean
	Uncommitted
	//	Detached
	//	Unpushed
	NoUpstream
	NotMaster
)

var (
	escapeWindows bool
	checked       bool
	mutex         sync.Mutex
)

type Context struct {
	output cmd.Output
}

func New(output cmd.Output) *Context {
	return &Context{output}
}

//Can't run this without a valid git repo, so defer it until we first see one
func (c *Context) checkEscape(path string, cmdContext *cmd.Context) {
	mutex.Lock()
	defer mutex.Unlock()
	if checked {
		return
	}

	//Some (but not all) versions of git on windows like the curly brackets
	//escaped. Rather than attempting to decipher this from the version of git
	//installed, I just retry the first time this runs with an escape, if that
	//works, then we switch over for remaining invokations

	_, err := cmdContext.Execf(`git rev-parse @{0}`)
	if err != nil {
		_, err := cmdContext.Execf(`git rev-parse @\{0\}`)
		if err != nil {
			errMsg := "Could not determine if git curly braces need to be escaped\n"
			errMsg += "git rev-parse @{0} failed both with and without quotes in"
			errMsg += path
			c.output.Error(errMsg)
			os.Exit(1)

		} else {
			escapeWindows = true
		}
	}

	checked = true

}

func escape(s string) string {
	if escapeWindows {
		left := strings.Replace(s, `{`, `\{`, -1)
		return strings.Replace(left, `}`, `\}`, -1)
	}
	return s
}

func (c *Context) Status(path string, must bool) (Status, error) {
	cmdContext := cmd.NewContext(path, c.output)
	var mustContext *cmd.Context
	if must {
		mustContext = cmd.NewContext(path, c.output, cmd.Must)
	} else {
		mustContext = cmdContext
	}

	c.checkEscape(path, cmdContext)

	if unchecked, err := mustContext.Execf("git status --porcelain"); err != nil {
		return Unknown, err
	} else if len(unchecked) != 0 {
		return Uncommitted, nil
	}

	if output, err := cmdContext.Execf("git branch --remote --contains"); err != nil {
		return Unknown, err
	} else {
		remoteBranches := dsutil.SplitLines(output, true)
		if len(remoteBranches) == 0 {
			return NoUpstream, nil
		}
		if !strings.Contains(output, "origin/master") {
			return NotMaster, nil
		}
	}
	return Clean, nil
}

func (c *Context) cmdContext(path string, must bool) *cmd.Context {
	if must {
		return cmd.NewContext(path, c.output, cmd.Must)
	} else {
		return cmd.NewContext(path, c.output, cmd.Warn)
	}
}

func (c *Context) silentContext(path string) *cmd.Context {
	return cmd.NewContext(path, c.output)
}

func (c *Context) TopLevel(path string, must bool) (string, error) {
	//Unfortunately you can't use "git rev-parse --show-toplevel", as it will unsymlinkify things.
	cmdContext := c.cmdContext(path, must)

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	sanePath, err := dsutil.SanitisePath(cmdContext, absPath)
	if err != nil {
		return "", err
	}

	splitPath := strings.Split(sanePath, "/")
	if runtime.GOOS == "windows" && len(splitPath[0]) == 2 && splitPath[0][1] == ':' {
		splitPath[0] = splitPath[0] + "\\"
	}
	for i := len(splitPath); i >= 0; i-- {
		tryPath := filepath.Join(splitPath[0:i]...)
		if dsutil.CheckPath(filepath.Join(tryPath, ".git")) {
			return dsutil.SanitisePath(cmdContext, tryPath)
		}

	}
	return "", fmt.Errorf("Could not find .git in any parent directory.")

}

func (c *Context) Clone(targetPath string, url string, must bool) error {
	err := os.MkdirAll(targetPath, 0755)
	if err != nil && must {
		c.output.Error("Could not create: %s", targetPath)
		os.Exit(1)
	} else if err != nil {
		return err
	}

	cmdContext := c.cmdContext(targetPath, must)

	_, err = cmdContext.Execf("git clone %s .", url)
	return err
}

func (c *Context) Checkout(targetPath, target string, must bool) error {
	cmdContext := c.cmdContext(targetPath, must)

	_, err := cmdContext.Execf("git checkout %s", target)
	return err
}

func (c *Context) Pull(targetPath string, must bool) error {
	cmdContext := c.cmdContext(targetPath, must)

	_, err := cmdContext.Execf("git pull")
	return err
}

func (c *Context) IsGit(targetPath string) bool {
	cmdContext := c.cmdContext(targetPath, false)

	_, err := cmdContext.Execf("git rev-parse")
	return err == nil
}

func (c *Context) SHA(targetPath string, must bool) (string, error) {
	cmdContext := c.cmdContext(targetPath, must)

	output, err := cmdContext.Execf("git rev-parse HEAD")
	return dsutil.FirstLine(output), err
}

func (c *Context) Tags(targetPath string, must bool) ([]string, error) {
	cmdContext := c.cmdContext(targetPath, must)

	output, err := cmdContext.Execf("git tag --points-at HEAD")
	return dsutil.SplitLines(output, true), err
}

func (c *Context) MostRecentTag(targetPath string, must bool) (string, error) {
	cmdContext := c.silentContext(targetPath)

	output, err := cmdContext.Execf("git describe --abbrev=0 --tags")
	if err != nil && strings.Contains(err.Error(), "No names found, cannot describe anything") {
		return "", nil
	} else if err != nil && must {
		panic(err)
	}

	return dsutil.FirstLine(output), err
}

func (c *Context) RemoteOriginUrl(targetPath string, must bool) (string, error) {
	cmdContext := c.cmdContext(targetPath, must)

	output, err := cmdContext.Execf("git config --get remote.origin.url")
	return dsutil.FirstLine(output), err
}

func (c *Context) AbbrevRef(targetPath string, must bool) (string, error) {
	cmdContext := c.cmdContext(targetPath, must)

	output, err := cmdContext.Execf("git rev-parse --abbrev-ref HEAD")
	return dsutil.FirstLine(output), err
}
