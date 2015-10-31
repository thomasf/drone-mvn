package main

import (
	"github.com/thomasf/drone-mvn/mavendeploy"
	"github.com/drone/drone-plugin-go/plugin"
)

func main() {
	workspace := plugin.Workspace{}
	repo := plugin.Repo{}
	build := plugin.Build{}
	vargs := mavendeploy.Maven{}

	plugin.Param("repo", &repo)
	plugin.Param("build", &build)
	plugin.Param("workspace", &workspace)
	plugin.Param("vargs", &vargs)
	plugin.MustParse()

	vargs.WorkspacePath(workspace.Path)

	err := vargs.Publish()
	if err != nil {
		panic(err)
	}
}
