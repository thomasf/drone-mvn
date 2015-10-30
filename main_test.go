package main

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestSkip(t *testing.T) {
	err := publish(&Maven{
		Repository: Repository{},
		Artifact:   Artifact{},
		GPG:        GPG{},
		Args:       Args{},
	}, "test-data/")

	if err != nil {
		t.Fatal(err)
	}

}

func TestURL(t *testing.T) {
	err := publish(&Maven{
		Repository: Repository{Username: "u", Password: "p"},
		Artifact:   Artifact{},
		GPG:        GPG{},
		Args:       Args{},
	}, "test-data/")
	if err == nil || err != errRequiredValue {
		t.Fatal("url should be required")
	}

}

func TestPublish1(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "drone-mvn-test")
	if err != nil {
		panic(err)
	}
	err = publish(&Maven{
		Repository: Repository{
			Username: "u",
			Password: "p",
			URL:      fmt.Sprintf("file://%s", tmpdir),
		},
		Artifact: Artifact{
			GroupID: "com.test.publish1",
		},
		GPG: GPG{},
		Args: Args{
			Source: "multiple-matched/app*",
			Regexp: "(?P<artifact>app-[^/-]*)-(?P<classifier>[^-]*-[^-]*)-(?P<version>.*).(?P<extension>tar.gz|zip|readme)$",
		},
	}, "test-data/")
	if err != nil {
		t.Fatal(err)
	}
}

func TestPublish2(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "drone-mvn-test")
	if err != nil {
		panic(err)
	}
	err = publish(&Maven{
		Repository: Repository{
			Username: "u",
			Password: "p",
			URL:      fmt.Sprintf("file://%s", tmpdir),
		},
		Artifact: Artifact{
			GroupID:    "com.test.publish2",
			ArtifactID: "release",
			Extension:  "zip",
			Version:    "1.2.3",
		},
		GPG: GPG{},
		Args: Args{
			Source: "single/release.zip",
		},
	}, "test-data/")
	if err != nil {
		t.Fatal(err)
	}
}

func TestPublish3(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "drone-mvn-test")
	if err != nil {
		panic(err)
	}
	err = publish(&Maven{
		Repository: Repository{
			Username: "u",
			Password: "p",
			URL:      fmt.Sprintf("file://%s", tmpdir),
		},
		Artifact: Artifact{
			GroupID:    "com.test.publish3",
			Extension:  "zip",
			Version:    "1.2.3",
		},
		GPG: GPG{},
		Args: Args{
			Source: "single-matched/*.zip",
			Regexp: "(?P<artifact>[^/-]*)-(?P<classifier>[^-]*-[^-]*).zip$",
		},
	}, "test-data/")
	if err != nil {
		t.Fatal(err)
	}
}
