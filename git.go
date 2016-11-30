package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/desal/cmd"
	"github.com/desal/dsutil"
	"github.com/desal/richtext"
)

//go:generate stringer -type Status
//go:generate stringer -type Flag

type (
	empty   struct{}
	Flag    int
	Status  int
	flagSet map[Flag]empty

	Context struct {
		format   richtext.Format
		cmdFlags []cmd.Flag
		flags    flagSet
	}
)

const (
	_ Status = iota
	Unknown
	Clean
	Uncommitted
	NoUpstream
	NotMaster

	_ Flag = iota
	MustExit
	MustPanic
	Warn
	Verbose
	LocalOnly // i.e. not being in remote is not considered an error
)

var (
	escapeWindows bool
	checked       bool
	mutex         sync.Mutex
	cmdFlags      = map[Flag]cmd.Flag{
		MustPanic: cmd.MustPanic,
		MustExit:  cmd.MustExit,
		Warn:      cmd.Warn,
		Verbose:   cmd.Verbose,
	}
)

func (fs flagSet) Checked(flag Flag) bool {
	_, ok := fs[flag]

	return ok
}

func New(format richtext.Format, flags ...Flag) *Context {
	c := &Context{format: format, flags: flagSet{}}

	c.cmdFlags = []cmd.Flag{cmd.TrimSpace}
	for _, flag := range flags {
		if cmdFlag, ok := cmdFlags[flag]; ok {
			c.cmdFlags = append(c.cmdFlags, cmdFlag)
		}
		c.flags[flag] = empty{}
	}
	return c
}

func (c *Context) errorf(s string, a ...interface{}) error {
	if c.flags.Checked(MustExit) {
		c.format.ErrorLine(s, a...)
		os.Exit(1)
	} else if c.flags.Checked(MustPanic) {
		panic(fmt.Errorf(s, a...))
	} else if c.flags.Checked(Warn) || c.flags.Checked(Verbose) {
		c.format.WarningLine(s, a...)
	}
	return fmt.Errorf(s, a...)
}

func (c *Context) warnf(s string, a ...interface{}) {
	if c.flags.Checked(Warn) {
		c.format.WarningLine(s, a...)
	}
}

func (c *Context) cmdContext(path string) *cmd.Context {
	return cmd.New(path, c.format, c.cmdFlags...)
}

func (c *Context) Status(path string) (Status, error) {
	ctx := c.cmdContext(path)

	if unchecked, _, err := ctx.Execf("git status --porcelain"); err != nil {
		return Unknown, err
	} else if len(unchecked) != 0 {
		return Uncommitted, nil
	}

	if c.flags.Checked(LocalOnly) {
		return Clean, nil
	}

	if output, _, err := ctx.Execf("git branch --remote --contains"); err != nil {
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

func (c *Context) TopLevel(path string) (string, error) {
	//Unfortunately you can't use "git rev-parse --show-toplevel", as it will unsymlinkify things.

	path = filepath.Clean(path)
	for {
		if dsutil.CheckPath(filepath.Join(path, ".git")) {
			return path, nil
		}

		nextPath := filepath.Join(path, "..")
		if nextPath == path {
			break
		}

		path = nextPath
	}

	return "", c.errorf("Could not find .git in any parent directory.")

}

func (c *Context) Clone(targetPath string, url string) error {
	err := os.MkdirAll(targetPath, 0755)
	if err != nil {
		return c.errorf("Could not create: %s", targetPath)
	}

	cmdContext := c.cmdContext(targetPath)

	_, _, err = cmdContext.Execf("git clone %s .", url)
	return err
}

func (c *Context) Checkout(targetPath, target string) error {
	cmdContext := c.cmdContext(targetPath)

	_, _, err := cmdContext.Execf("git checkout %s", target)
	return err
}

func (c *Context) Pull(targetPath string) error {
	cmdContext := c.cmdContext(targetPath)

	_, _, err := cmdContext.Execf("git pull")
	return err
}

func (c *Context) IsGit(targetPath string) bool {
	cmdContext := c.cmdContext(targetPath)

	stdout, _, _ := cmdContext.Execf("git rev-parse --is-inside-work-tree; exit 0")
	return strings.Contains(stdout, "true")
}

func (c *Context) SHA(targetPath string) (string, error) {
	cmdContext := c.cmdContext(targetPath)

	output, _, err := cmdContext.Execf("git rev-parse HEAD")
	return dsutil.FirstLine(output), err
}

func (c *Context) Tags(targetPath string) ([]string, error) {
	cmdContext := c.cmdContext(targetPath)

	output, _, err := cmdContext.Execf("git tag --points-at HEAD")
	return dsutil.SplitLines(output, true), err
}

func (c *Context) MostRecentTag(targetPath string) (string, error) {
	cmdContext := c.cmdContext(targetPath)

	stdout, stderr, err := cmdContext.Execf("git describe --abbrev=0 --tags; exit 0")
	if strings.Contains(stderr, "No names found, cannot describe anything") ||
		strings.Contains(stderr, "No tags can describe") {
		return "", nil
	} else if stderr != "" {
		_, _, err := cmdContext.Execf("git describe --abbrev=0 --tags")
		return "", err
	}

	return dsutil.FirstLine(stdout), err
}

func (c *Context) RemoteOriginUrl(targetPath string) (string, error) {
	cmdContext := c.cmdContext(targetPath)

	output, _, err := cmdContext.Execf("git config --get remote.origin.url")
	return dsutil.FirstLine(output), err
}

func (c *Context) AbbrevRef(targetPath string) (string, error) {
	cmdContext := c.cmdContext(targetPath)

	output, _, err := cmdContext.Execf("git rev-parse --abbrev-ref HEAD")
	return dsutil.FirstLine(output), err
}
