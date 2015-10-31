Use the mvn plugin to upload files and build artifacts to maven repositories.

Maven property options:

* **username** - maven username
* **password** - maven password
* **group_id** - default artifact group ID
* **artifact_id** - default artifact ID
* **version** - default artifact version
* **classifier** - default artifact classifier
* **extension** - default artifact extension

Drone-mvn maven options:

* **source** - location of files to upload (supports globbing)
* **regexp** - regexp with named groups to parse globbed files into maven artifacts, the maven property options above are used as defaults if the regexp doesnt contain one or more of the properites. See the drone-mvn [tests](https://github.com/thomasf/drone-mvn/blob/694f52340274f3c6304aaa678bcead27761fcb76/mavendeploy/mavendeploy_test.go#L55) for some examples of source/regexp interaction. The valid regexp capturing groups are **version**, **classifier**,  **artifact**,  **group** and **extension**.

gpg signing options:

* **gpg_private_key** - in gnupg private key pem format
* **gpg_passphrase** - in clear text


The following is a sample mvn configuration in your .drone.yml file:

```yaml
publish:
  mvn:
    username: "user"
    password: "mypassword"
```
