[![Build Status](https://drone.23c.se/api/badges/thomasf/drone-mvn/status.svg)](https://drone.23c.se/thomasf/drone-mvn)

# drone-mvn

`drone-mvn` is a publsher plugin for [Drone CI](https://github.com/drone/drone).

## Drone publisher: drone-mvn

- See [DOCS.md](https://github.com/thomasf/drone-mvn/blob/master/DOCS.md) for
  how to use as a publisher in drone.

## Docker image [@Docker Hub](https://hub.docker.com/r/thomasf/drone-mvn/)

Tags (soon):

 - thomasf/drone-mvn <- latest version
 - thomasf/drone-mvn:latest <- latest version
 - thomasf/drone-mvn:[version] <- semver 2.0 entry from git tag
 - thomasf/drone-mvn:master <- latest master branch commit

## Source code  [@GitHub](https://github.com/thomasf/drone-mvn)

Go wrapper around the Maven ang GnuPG command line tools.

- godoc: https://godoc.org/github.com/thomasf/drone-mvn/mavendeploy
- main tests: https://github.com/thomasf/drone-mvn/blob/master/main_test.go
- mavendeploy tests: https://github.com/thomasf/drone-mvn/blob/master/mavendeploy/mavendeploy_test.go
