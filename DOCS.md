# Drone-mvn

This plugin is at the time of writing only known to work with drone 0.4
which is in beta. The example configuration below assumes that you are running
your own drone instance and have
[configured it correctly](http://readme.drone.io/setup/plugins.html) to allow
this plugin.

## Introduction

Use the mvn plugin to upload files and build artifacts to maven repositories.

`drone-mvn` is mainly targeted at publishing builds of non Java projects since
maven/gradle can handle those fine. Sonatype Nexus (OSS) is a fine solution for
storing build artifacts of any type. Nexus has a very configurable
authorization system which allows for giving access to artifacts by path
matching and much much more.

## options for .drone.yml

Maven property options:

* **username** - maven username
* **password** - maven password
* **group** - default artifact group ID
* **artifact** - default artifact ID
* **version** - default artifact version
* **classifier** - default artifact classifier
* **extension** - default artifact extension

Drone-mvn maven options:

* **source** - location of files to upload (supports globbing)
* **regexp** - regexp with named groups to parse globbed files into maven artifacts, the maven property options above are used as defaults if the regexp doesnt contain one or more of the properites. See the drone-mvn [tests](https://github.com/thomasf/drone-mvn/blob/694f52340274f3c6304aaa678bcead27761fcb76/mavendeploy/mavendeploy_test.go#L55) for some examples of source/regexp interaction. The valid regexp capturing groups are **version**, **classifier**,  **artifact**,  **group** and **extension**.

gpg signing options:

* **gpg_private_key** - in gnupg private key pem format
* **gpg_passphrase** - in clear text

## Configuration examples

An example of .drone.yml publish configuration of a single snapshot or release build artifact:

```yaml
publish:
  drone-mvn:
    image: thomasf/drone-mvn
    username: my-maven-username
    password: my-maven-password
    url: https://nexus.mycompany.com/content/repositories/project-snapshots/
    group: com.mycompany.project
    artifact: webassets
    version: SNAPSHOT
    source: release/web*.tgz
    extension: tgz

    when:
      branch: master

  drone-mvn:
    image: thomasf/drone-mvn
    username: my-maven-username
    password: my-maven-password
    url: https://nexus.mycompany.com/content/repositories/project-releases/
    group: com.mycompany.project
    artifact: webassets
    version: $$TAG
    source: release/web*.tgz
    extension: tgz
    
    when:
        event: tag
```

