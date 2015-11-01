package main

import (
	"flag"
	"os"

	"github.com/drone/drone-plugin-go/plugin"
	"github.com/thomasf/drone-mvn/mavendeploy"
)

// testExpressions allows for quickly testing source/regexp patterns against
// files via the command line.
func testExpressions() {
	var regexp, source string
	flag.StringVar(&regexp, "regexp", "", "regular expression to test")
	flag.StringVar(&source, "source", "", "source expression to test")
	flag.Parse()
	if regexp != "" && source != "" {
		mvn := mavendeploy.Maven{
			Artifact: mavendeploy.Artifact{
				GroupID:    "GROUPID",
				ArtifactID: "ARTIFACTID",
				Version:    "99.99.99",
				Extension:  "EXTENSION",
			},
			Args: mavendeploy.Args{
				Debug:  true,
				Source: source,
				Regexp: regexp,
			},
		}
		mvn.WorkspacePath(".")
		err := mvn.Prepare()
		if err != nil {
			panic(err)
		}
		os.Exit(0)
	}
}

func main() {
	testExpressions()
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
