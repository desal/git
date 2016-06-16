package main

import (
	"fmt"

	"github.com/desal/cmd"
	"github.com/desal/git"
	"github.com/desal/richtext"
)

func main() {
	output := cmd.NewStdOutput(false, richtext.Ansi())
	ctx := git.New(output)

	fmt.Println("IsGit:", ctx.IsGit("."))

	url, err := ctx.RemoteOriginUrl(".", false)
	fmt.Println("RemoteOriginUrl:", url, err)

	sha, err := ctx.SHA(".", false)
	fmt.Println("SHA:", sha, err)

	tags, err := ctx.Tags(".", false)
	fmt.Println("Tags:", tags, err)

	mostRecentTag, err := ctx.MostRecentTag(".", false)
	fmt.Println("MostRecentTag:", mostRecentTag, err)

	status, err := ctx.Status(".", false)
	fmt.Println("Status:", status, err)

	topLevel, err := ctx.TopLevel(".", false)
	fmt.Println("TopLevel:", topLevel, err)

}
