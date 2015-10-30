Use the mvn plugin to upload files and build artifacts to maven repositories.

* **username** - maven username
* **password** - maven password

The following is a sample mvn configuration in your .drone.yml file:

```yaml
publish:
  mvn:
    username: "user"
    password: "mypassword"
```
