package mavendeploy

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

	"github.com/davecgh/go-spew/spew"
)

const (
	mavenGpg    = "org.apache.maven.plugins:maven-gpg-plugin:1.6:sign-and-deploy-file"
	mavenDeploy = "org.apache.maven.plugins:maven-deploy-plugin:2.8.2:deploy-file"
)

// Maven is a composed struct which forms the configration of the drone-mvn
// drone plugin.
type Maven struct {
	Repository // maven repository
	Artifact   // artifact
	GPG        // signing information
	Args       // drone-mvn specific options

	gpgCmd        *GpgCmd
	workspacePath string
	settingsPath  string
	artifacts     map[string][]Artifact
	quiet         bool
}

// Repository is a target Maven repository configuration
type Repository struct {
	Username string `json:"username"`
	Password string `json:"password"`
	URL      string `json:"url"`
}

// Artifact is a target Maven artifact.
type Artifact struct {
	GroupID    string `json:"group"`      // e.g. org.springframework
	ArtifactID string `json:"artifact"`   // e.g. spring-core
	Version    string `json:"version"`    // e.g. 4.1.3.RELEASE
	Classifier string `json:"classifier"` // e.g. sources, javadoc, <the empty string>...
	Extension  string `json:"extension"`  // e.g. jar, .tar.gz, .zip
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

var (
	errRequiredValue = errors.New("required")
	errInvalidValue  = errors.New("invalid")
	errNotFound      = errors.New("not found")
)

func (mvn *Maven) WorkspacePath(path string) error {
	mvn.workspacePath = path
	return nil
}

func (mvn *Maven) Publish() error {
	if mvn.quiet {
		mvn.Args.Debug = false
	}
	// skip if Repository Username or Password are empty. A good example for
	// this would be forks building a project.
	if mvn.Repository.Username == "" || mvn.Repository.Password == "" {
		mvn.infof("username or password is empty, skipping publish")
		return nil
	}
	if mvn.Repository.URL == "" {
		mvn.infof("URL is not set")
		return errRequiredValue
	}

	err := mvn.Prepare()
	if err != nil {
		return err
	}
	if mvn.GPG.PrivateKey != "" {
		gpgCmd := &GpgCmd{GPG: mvn.GPG}
		err := gpgCmd.Setup()
		if err != nil {
			return err
		}
		defer func() {
			err := gpgCmd.Teardown()
			if err != nil {
				panic(err)
			}
		}()
		mvn.gpgCmd = gpgCmd
	}
	settings, err := m2Settings(*mvn)
	if err != nil {
		return err
	}

	mvn.settingsPath = settings
	if !mvn.quiet {
		fmt.Println("$", settings)
	}
	defer func() {

		os.Remove(settings)
	}()
	var commands []*exec.Cmd
	for _, v := range mvn.artifacts {
		cmd := mvn.command(v...)
		cmd.Env = os.Environ()
		if !mvn.quiet {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
		commands = append(commands, cmd)
	}
	for _, cmd := range commands {
		if !mvn.quiet {
			mvn.trace(cmd)
		}
		err = cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

func (mvn *Maven) Prepare() error {
	sources, err := filepath.Glob(mvn.workspacePath + string(os.PathSeparator) + mvn.Args.Source)
	if err != nil {
		return err
	}
	if len(sources) == 0 {
		return fmt.Errorf("no sources found for %s ", mvn.Args.Source)
	}
	if mvn.Args.Debug {
		fmt.Println("sources found:")
		spew.Dump(sources)
	}
	if len(sources) > 1 {
		if mvn.Args.Regexp == "" {
			return fmt.Errorf(
				"multiple sources found for %s (%v) but no regexp was defined",
				mvn.Args.Source, sources)
		}
	}

	var parsed []Artifact
	if mvn.Args.Regexp == "" {
		a := mvn.Artifact
		a.file = sources[0]
		parsed = append(parsed, a)
	} else {
		re, err := regexp.Compile(mvn.Args.Regexp)
		if err != nil {
			return err
		}
		for _, s := range sources {
			rel, err := filepath.Rel(mvn.workspacePath, s)
			if err != nil {
				fmt.Printf("could not make source %s relative to %s\n", s, mvn.workspacePath)
				return err
			}
			matches := re.FindStringSubmatch(rel)
			if matches == nil {
				return fmt.Errorf("regexp '%s' does not match '%s'", mvn.Args.Regexp, s)
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
			if mvn.Args.Debug {
				fmt.Println("$ parsed artifact")
				spew.Dump(a)
			}
		}
		if len(parsed) == 0 {
			return errNotFound
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
			a.GroupID = mvn.Artifact.GroupID
		}
		if a.Version == "" {
			a.Version = mvn.Artifact.Version
		}
		if a.ArtifactID == "" {
			a.ArtifactID = mvn.Artifact.ArtifactID
		}
		if a.Classifier == "" {
			a.Classifier = mvn.Artifact.Classifier
		}
		if a.Extension == "" {
			a.Extension = mvn.Artifact.Extension
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

	if len(mapped) == 0 {
		return errNotFound
	}

	mvn.artifacts = mapped
	if mvn.Args.Debug {
		fmt.Println("mapped artifacts")
		spew.Dump(mapped)
	}

	return nil
}

// command is a helper function that returns the command
// and arguments to upload to aws from the command line.
func (mvn Maven) command(artifacts ...Artifact) *exec.Cmd {

	var args []string
	args = append(args, "-B")

	switch {
	case mvn.quiet:
		args = append(args, "-q")
	case mvn.Debug:
		args = append(args, "-X")
	}

	args = append(args,
		"--settings", mvn.settingsPath,
	)
	if mvn.gpgCmd != nil {

		args = append(args,
			mavenGpg,
			fmt.Sprintf("-Dgpg.defaultKeyring=false"),
			fmt.Sprintf("-Dgpg.publicKeyring=%s", mvn.gpgCmd.PublicRing),
			fmt.Sprintf("-Dgpg.secretKeyring=%s", mvn.gpgCmd.SecretRing),
			fmt.Sprintf("-Dgpg.keyname=%s", mvn.gpgCmd.SecretKeyID),
			fmt.Sprintf("-Dgpg.passphraseServerId=%s", gpgServerID),
			fmt.Sprintf("-Dgpg.ascDirectory=%s", mvn.gpgCmd.tempDir),
		)
	} else {
		args = append(args, mavenDeploy)
	}

	a := artifacts[0]
	args = append(args,
		fmt.Sprintf("-Durl=%s", mvn.Repository.URL),
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

// trace writes each command to standard error (preceded by a ‘$ ’) before it
// is executed. Used for debugging your build.
func (mvn *Maven) trace(cmd *exec.Cmd) {
	if !mvn.quiet {
		fmt.Println("$", strings.Join(cmd.Args, " "))
	}
}

func (mvn *Maven) infof(format string, a ...interface{}) {
	if !mvn.quiet {
		fmt.Println("$", fmt.Sprintf(format, a...))
	}
}

// static id's for maven repo id and gpg auth info generation.
const (
	deployRepoID = "deploy-repo"
	gpgServerID  = "gpg-auth"
)
