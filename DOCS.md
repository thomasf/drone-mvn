**Introduction**

Use the mvn plugin to upload files and build artifacts to maven repositories.

`drone-mvn` **is most useful for publishing builds of non Java projects** since
Gradle and Maven usually handles Java project publishing without needing any
additional help.

The artifact repository manager
[Sonatype Nexus (OSS)](http://www.sonatype.org/nexus/) is a great solution for
storing build artifacts of any type.

**Notice**

This plugin is at the time of writing only known to work with
[drone 0.4](https://github.com/drone/drone) which is in beta.

The example configurations below assumes that **you are running your own
drone** instance and have
[configured it correctly](http://readme.drone.io/setup/plugins.html) to allow
this plugin.

The **arguments for .drone.yml** are **probably final**. I might still change how
publishing is specified but not unless there is a good enough reason to break
usage.


**Options for .drone.yml**

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

GnuPG signing options:

* **gpg_private_key** - in gnupg private key pem format
* **gpg_passphrase** - in clear text

**Links**

- [GitHub](https://github.com/thomasf/drone-mvn)
- [Docker Hub](https://hub.docker.com/r/thomasf/drone-mvn/))

**Configuration examples**

**An example of .drone.yml publish configuration of a single snapshot or release build artifact:**

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


**An example where artifacts are mapped by globbing and regular expression:**

In this example the build system has created the following files which is to be
published by the `drone-mvn` publisher:

```
build/archives/
├── app-client-darwin-amd64-0.1.4.zip
├── app-client-linux-386-0.1.4.tar.gz
├── app-client-linux-amd64-0.1.4.tar.gz
├── app-client-windows-386-0.1.4.zip
├── app-client-windows-amd64-0.1.4.zip
├── app-gui-darwin-amd64-0.1.4.zip
├── app-server-linux-amd64-0.1.4.readme
├── app-server-linux-amd64-0.1.4.tar.gz
└── README.md
```

Inside .drone.yml

```
publish:
  drone-mvn:
    image: thomasf/drone-mvn
    username: $$NEXUS_USER
    password: $$NEXUS_PASSWORD
    url: https://nexus.mycompany.com/content/repositories/myproject-releases/
    group: com.test.publish1
    version: $$TAG
    source: build/archives/*
    regexp: "(?P<artifact>app-[^/-]*)-(?P<classifier>[^-]*-[^-]*)-.*(?P<extension>tar.gz|zip|readme)$"

    when:
      event: tag

```

The resulting maven artifacts becomes

```
com/
└── test
    └── publish1
        ├── app-client
        │   ├── 0.1.4
        │   │   ├── app-client-0.1.4-darwin-amd64.zip.md5
        │   │   ├── app-client-0.1.4-darwin-amd64.zip.sha1
        │   │   ├── app-client-0.1.4-linux-386.tar.gz.md5
        │   │   ├── app-client-0.1.4-linux-386.tar.gz.sha1
        │   │   ├── app-client-0.1.4-linux-amd64.tar.gz.md5
        │   │   ├── app-client-0.1.4-linux-amd64.tar.gz.sha1
        │   │   ├── app-client-0.1.4.pom
        │   │   ├── app-client-0.1.4.pom.md5
        │   │   ├── app-client-0.1.4.pom.sha1
        │   │   ├── app-client-0.1.4-windows-386.zip.md5
        │   │   ├── app-client-0.1.4-windows-386.zip.sha1
        │   │   ├── app-client-0.1.4-windows-amd64.zip.md5
        │   │   └── app-client-0.1.4-windows-amd64.zip.sha1
        │   ├── maven-metadata.xml
        │   ├── maven-metadata.xml.md5
        │   └── maven-metadata.xml.sha1
        ├── app-gui
        │   ├── 0.1.4
        │   │   ├── app-gui-0.1.4-darwin-amd64.zip.md5
        │   │   ├── app-gui-0.1.4-darwin-amd64.zip.sha1
        │   │   ├── app-gui-0.1.4.pom
        │   │   ├── app-gui-0.1.4.pom.md5
        │   │   └── app-gui-0.1.4.pom.sha1
        │   ├── maven-metadata.xml
        │   ├── maven-metadata.xml.md5
        │   └── maven-metadata.xml.sha1
        └── app-server
            ├── 0.1.4
            │   ├── app-server-0.1.4-linux-amd64.readme.md5
            │   ├── app-server-0.1.4-linux-amd64.readme.sha1
            │   ├── app-server-0.1.4-linux-amd64.tar.gz.md5
            │   ├── app-server-0.1.4-linux-amd64.tar.gz.sha1
            │   ├── app-server-0.1.4.pom
            │   ├── app-server-0.1.4.pom.md5
            │   └── app-server-0.1.4.pom.sha1
            ├── maven-metadata.xml
            ├── maven-metadata.xml.md5
            └── maven-metadata.xml.sha1

8 directories, 34 files
```
