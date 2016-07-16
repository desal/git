package main

import (
	"fmt"

	"github.com/desal/git"
	"github.com/desal/richtext"
)

func main() {
	format := richtext.New()
	ctx := git.New(format)

	fmt.Println("IsGit:", ctx.IsGit("."))

	url, err := ctx.RemoteOriginUrl(".")
	fmt.Println("RemoteOriginUrl:", url, err)

	sha, err := ctx.SHA(".")
	fmt.Println("SHA:", sha, err)

	tags, err := ctx.Tags(".")
	fmt.Println("Tags:", tags, err)

	mostRecentTag, err := ctx.MostRecentTag(".")
	fmt.Println("MostRecentTag:", mostRecentTag, err)

	status, err := ctx.Status(".")
	fmt.Println("Status:", status, err)

	topLevel, err := ctx.TopLevel(".")
	fmt.Println("TopLevel:", topLevel, err)

}
