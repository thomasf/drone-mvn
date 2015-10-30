package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/drone/drone-plugin-go/plugin"
)

const (
	mavenGpg    = "org.apache.maven.plugins:maven-gpg-plugin:1.6:sign-and-deploy-file"
	mavenDeploy = "org.apache.maven.plugins:maven-deploy-plugin:2.7:deploy-file"
)

// Maven is a composed struct which forms the configration of the drone-mvn
// drone plugin.
type Maven struct {
	Repository // maven repository
	Artifact   // artifact
	GPG        // signing information
	Args       // drone-mvn specific options

	workspacePath string
	artifacts     map[string][]Artifact
}

// Debug enabled verbose logging
var Debug = false

// Repository is a target Maven repository configuration
type Repository struct {
	Username string `json:"username"`
	Password string `json:"password"`
	URL      string `json:"url"`
}

// Artifact is a target Maven artifact.
type Artifact struct {
	GroupID    string `json:"group_id"`    // e.g. org.springframework
	ArtifactID string `json:"artifact_id"` // e.g. spring-core
	Version    string `json:"version"`     // e.g. 4.1.3.RELEASE
	Classifier string `json:"classifier"`  // e.g. sources, javadoc, <the empty string>...
	Extension  string `json:"extension"`   // e.g. jar, .tar.gz, .zip
	file       string
}

// Args is the drone-mvn specific arguments.
// If there are multiple matches to Source, ArtifactRegexp must be defined.
type Args struct {
	Source string `json:"source"` // artifact filename glob
	Regexp string `json:"regexp"` // parses artifact filenames to artifacts
	Debug  bool   `json:"debug"`  // debug output
}

// GPG holds the GnuPG key information used for signing releases.
type GPG struct {
	PrivateKey string `json:"gpg_private_key"` // private key
	Passphrase string `json:"gpg_passphrase"`  // private key passphrase (optional)
}

func main() {
	workspace := plugin.Workspace{}
	repo := plugin.Repo{}
	build := plugin.Build{}
	vargs := Maven{}

	plugin.Param("repo", &repo)
	plugin.Param("build", &build)
	plugin.Param("workspace", &workspace)
	plugin.Param("vargs", &vargs)
	plugin.MustParse()

	err := publish(&vargs, workspace.Path)
	if err != nil {
		panic(err)
	}
}

var (
	errRequiredValue = errors.New("required")
	errInvalidValue  = errors.New("invalid")
)

func publish(mvn *Maven, workspacePath string) error {

	mvn.workspacePath = workspacePath
	if !Debug {
		Debug = mvn.Args.Debug
	}

	// skip if Repository Username or Password are empty. A good example for
	// this would be forks building a project.
	if mvn.Repository.Username == "" || mvn.Repository.Password == "" {
		fmt.Println("username or password is empty, skipping publish")
		return nil
	}

	if mvn.Repository.URL == "" {
		fmt.Println("URL is not set")
		return errRequiredValue
	}

	err := mvn.parseSources()
	if err != nil {
		return err
	}

	settings, err := m2Settings(*mvn)
	if err != nil {
		return err
	}
	fmt.Println("$", settings)
	defer func() {
		os.Remove(settings)
	}()

	var commands []*exec.Cmd
	for _, v := range mvn.artifacts {
		cmd := command(settings, mvn.Repository, mvn.GPG, v...)
		cmd.Env = os.Environ()
		// cmd.Dir = workspacePath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		commands = append(commands, cmd)
	}

	for _, cmd := range commands {
		trace(cmd)

		// run the command and exit if failed.
		err = cmd.Run()
		if err != nil {
			return err
		}

	}
	return nil
}

func (m *Maven) parseSources() error {
	sources, err := filepath.Glob(m.workspacePath + string(os.PathSeparator) + m.Args.Source)
	if err != nil {
		return err
	}

	if len(sources) == 0 {
		return fmt.Errorf("no sources found for %s ", m.Args.Source)
	}

	if len(sources) > 1 {
		if m.Args.Regexp == "" {
			return fmt.Errorf(
				"multiple sources found for %s (%v) but no regexp was defined",
				m.Args.Source, sources)
		}
	}

	var parsed []Artifact
	if m.Args.Regexp == "" {
		parsed = append(parsed, m.Artifact)
	} else {
		re, err := regexp.Compile(m.Args.Regexp)
		if err != nil {
			return err
		}
		for _, s := range sources {
			matches := re.FindStringSubmatch(s)
			if matches == nil {
				return fmt.Errorf("regexp '%s' does not match '%s'", m.Args.Regexp, s)
			}
			var a Artifact
			for i, name := range re.SubexpNames() {
				v := matches[i]
				switch name {
				case "":
					continue
				case "version":
					a.Version = v
				case "classifier":
					a.Classifier = v
				case "artifact":
					a.ArtifactID = v
				case "group":
					a.GroupID = v
				case "extension":
					a.Extension = v
				default:
					return fmt.Errorf("key %s not recognized by drone-mvn", name)
				}
			}
			a.file = s
			parsed = append(parsed, a)
			if Debug {
				fmt.Printf("$ parsed artifact: %v\n", a)
			}

		}

		// partition parsed artifacts into a map
		mapped := make(map[string][]Artifact, 0)
		mapkey := func(a Artifact) string {
			return fmt.Sprintf("%s:%s:%s", a.GroupID, a.ArtifactID, a.Version)
		}
		fill := func(orig Artifact) Artifact {
			a := orig
			if a.GroupID == "" {
				a.GroupID = m.Artifact.GroupID
			}
			if a.Version == "" {
				a.Version = m.Artifact.Version
			}
			if a.ArtifactID == "" {
				a.ArtifactID = m.Artifact.ArtifactID
			}
			return a
		}
		for _, v := range parsed {
			filled := fill(v)
			key := mapkey(filled)
			var artifacts []Artifact
			if _, ok := mapped[key]; ok {
				artifacts = mapped[key]
			}
			artifacts = append(artifacts, filled)
			mapped[key] = artifacts
		}

		m.artifacts = mapped
	}

	return nil
}

// command is a helper function that returns the command
// and arguments to upload to aws from the command line.
func command(settingspath string, repo Repository, gpg GPG, artifacts ...Artifact) *exec.Cmd {

	var args []string
	if Debug {
		args = append(args, "-X")
	} else {
		args = append(args, "-q")
	}

	args = append(args,
		"--settings", settingspath,
	)

	if gpg.PrivateKey != "" {
		fmt.Println("WARNING: GPG signing is not yet implmented")
		args = append(args,
			mavenGpg,
			fmt.Sprintf("-Dgpg.passphraseServerId=%s", gpgServerID),
			fmt.Sprintf("-Dgpg.keyname=%s", "TODO"),
		)
	} else {
		args = append(args, mavenDeploy)
	}

	a := artifacts[0]
	args = append(args,
		fmt.Sprintf("-Durl=%s", repo.URL),
		fmt.Sprintf("-DrepositoryId=%s", deployRepoID),
		fmt.Sprintf("-DgroupId=%s", a.GroupID),
		fmt.Sprintf("-DartifactId=%s", a.ArtifactID),
		fmt.Sprintf("-Dversion=%s", a.Version),
		fmt.Sprintf("-Dfile=%s", a.file),
	)
	if a.Extension != "" {
		args = append(args, fmt.Sprintf("-Dpackaging=%s", a.Extension))
	}
	if a.Classifier != "" {
		args = append(args, fmt.Sprintf("-Dclassifier=%s", a.Classifier))
	}

	if len(artifacts) > 1 {
		var files, types, classifiers []string
		for _, v := range artifacts[1:] {
			files = append(files, v.file)
			types = append(types, v.Extension)
			classifiers = append(classifiers, v.Classifier)
		}
		args = append(args, fmt.Sprintf("-Dfiles=%s", strings.Join(files, ",")))
		args = append(args, fmt.Sprintf("-Dclassifiers=%s", strings.Join(classifiers, ",")))
		args = append(args, fmt.Sprintf("-Dtypes=%s", strings.Join(types, ",")))
	}

	return exec.Command("mvn", args...)
}

// trace writes each command to standard error (preceded by a ‘$ ’) before it
// is executed. Used for debugging your build.
func trace(cmd *exec.Cmd) {
	fmt.Println("$", strings.Join(cmd.Args, " "))
}

// Settings is the root of the maven settings.xml file
type Settings struct {
	XMLName xml.Name `xml:"settings"`
	Servers []Server `xml:"servers>server"`
}

// Server entry for the maven settings.xml
type Server struct {
	ID         string `xml:"id"`
	Username   string `xml:"username,omitempty"`
	Password   string `xml:"password,omitempty"`
	Passphrase string `xml:"passphrase,omitempty"`
}

func m2Settings(m Maven) (string, error) {
	var servers []Server
	servers = append(servers, Server{
		ID:       deployRepoID,
		Username: m.Repository.Username,
		Password: m.Repository.Password,
	})
	if m.GPG.PrivateKey != "" {
		servers = append(servers, Server{
			ID:         gpgServerID,
			Passphrase: m.GPG.Passphrase,
		})
	}

	settings := Settings{
		Servers: servers,
	}

	output, err := xml.MarshalIndent(settings, "", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return "", err
	}

	f, err := ioutil.TempFile("", "drone-mvn-settings")
	if err != nil {
		return "", err
	}

	_, err = f.Write(output)
	if err != nil {
		os.Remove(f.Name())
		return "", err
	}
	f.Close()
	return f.Name(), nil

}

// static id's for maven repo id and gpg auth info generation.
const (
	deployRepoID = "deploy-repo"
	gpgServerID  = "gpg-auth"
)