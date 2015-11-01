package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"
	"testing"
	"text/template"
)

const testTemplate1 = `{
    "repo" : {
        "owner": "foo",
        "name": "bar",
        "full_name": "foo/bar"
    },
    "system": {
        "link_url": "http://drone.mycompany.com"
    },
    "workspace": {
        "path": "mavendeploy/"
    },
    "build" : {
        "number": 22,
        "status": "success",
        "started_at": 1421029603,
        "finished_at": 1421029813,
        "commit": "9f2849d5",
        "branch": "master",
        "message": "Update the Readme",
        "author": "johnsmith",
        "author_email": "john.smith@gmail.com"
    },
    "vargs": {
        "username": "someuser",
        "password": "somepassword",
        "url": "{{.URL}}",
        "group": "com.alkasir.test",
        "artifact": "Dockerfile",
        "packaging": "Dockerfile",
        "source": "test-data/multiple-matched/app*",
        "regexp": "(?P<artifact>app-[^-]*)-(?P<classifier>[^-]*-[^-]*)-(?P<version>.*).(?P<extension>tar.gz|zip|readme)"
    }
}`

func TestPlugin(t *testing.T) {
	if os.Getenv("__TEST_SUBCMD") == "1" {
		main()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestPlugin")
	env := append(os.Environ(), "__TEST_SUBCMD=1")
	cmd.Env = env
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := cmd.Run()
		if err != nil {
			t.Fatal(err)
		}
	}()
	tpl := template.Must(template.New("template").Parse(testTemplate1))

	tmpdir, err := ioutil.TempDir("", "drone-mvn-main-test")
	if err != nil {
		panic(err)
	}
	defer func() {
		os.RemoveAll(tmpdir)
	}()
	URL := fmt.Sprintf("file://%s", tmpdir)
	err = tpl.Execute(stdin, &struct{ URL string }{URL: URL})
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()
}
