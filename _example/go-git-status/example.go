package main

import (
	"fmt"
	"time"

	"github.com/desal/git"
	"github.com/desal/richtext"
)

func timeIt(fn func()) {
	tBefore := time.Now()
	fn()
	tAfter := time.Now()
	fmt.Println(" in ", tAfter.Sub(tBefore))
}

func main() {
	format := richtext.New()
	ctx := git.New(format)

	timeIt(func() {
		fmt.Print("IsGit: ", ctx.IsGit("."))
	})

	timeIt(func() {
		url, err := ctx.RemoteOriginUrl(".")
		fmt.Print("RemoteOriginUrl: ", url, err)
	})

	timeIt(func() {
		sha, err := ctx.SHA(".")
		fmt.Print("SHA: ", sha, err)
	})

	timeIt(func() {
		ct, err := ctx.CommitTime(".")
		fmt.Print("CommitTime: ", ct, err)
	})

	timeIt(func() {
		tags, err := ctx.Tags(".")
		fmt.Print("Tags: ", tags, err)
	})

	timeIt(func() {
		mostRecentTag, err := ctx.MostRecentTag(".")
		fmt.Print("MostRecentTag: ", mostRecentTag, err)
	})

	timeIt(func() {
		status, err := ctx.Status(".")
		fmt.Print("Status: ", status, err)
	})

	timeIt(func() {
		topLevel, err := ctx.TopLevel(".")
		fmt.Print("TopLevel: ", topLevel, err)
	})

}
