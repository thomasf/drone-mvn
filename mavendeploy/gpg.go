package mavendeploy

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// GpgCmd wraps the GnuPG command line util to create a temporary keychain for
type GpgCmd struct {
	GPG GPG

	tempDir     string
	PublicRing  string
	SecretRing  string
	SecretKeyID string
	Quiet       bool
}

func (g *GpgCmd) Setup() error {
	tmpdir, err := ioutil.TempDir("", "drone-mvn-keydir")
	if err != nil {
		return err
	}
	g.tempDir = tmpdir
	g.SecretRing = filepath.Join(tmpdir, "secret.gpg")
	g.PublicRing = filepath.Join(tmpdir, "public.gpg")
	err = g.importKeys()
	if err != nil {
		g.Teardown()
		return err
	}
	return nil
}

func (g *GpgCmd) Teardown() error {
	if g.tempDir != "" {
		return os.RemoveAll(g.tempDir)
	}
	return nil
}

func (g *GpgCmd) newCmd(args ...string) *exec.Cmd {
	var cmdArgs []string
	cmdArgs = append(cmdArgs,
		"--quiet",
		"--no-default-keyring",
		fmt.Sprintf("--secret-keyring=%s", g.SecretRing),
		fmt.Sprintf("--keyring=%s", g.PublicRing),
	)
	cmdArgs = append(cmdArgs, args...)
	return exec.Command("gpg", cmdArgs...)
}

func (g *GpgCmd) importKeys() error {

	// import private key from pem string
	{
		cmd := g.newCmd("--import")
		if !g.Quiet {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return err
		}
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := cmd.Run()
			if err != nil {
				panic(err)
			}

		}()
		stdin.Write([]byte(g.GPG.PrivateKey))
		stdin.Close()
		wg.Wait()
	}

	// find the default secret key id
	{
		cmd := g.newCmd("--list-secret-keys", "--with-colons")
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		if !g.Quiet {
			cmd.Stderr = os.Stderr
		}
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := cmd.Run()
			if err != nil {
				panic(err)
			}

		}()

		out, err := ioutil.ReadAll(stdout)
		if err != nil {
			return err
		}
		lines := strings.Split(string(out), "\n")

	loop:
		for _, v := range lines {
			item := strings.Split(v, ":")
			if len(item) < 5 {
				return fmt.Errorf("line '%s' has too few colons", v)
			}
			if item[0] == "sec" {
				g.SecretKeyID = item[4]
				break loop
			}
		}
		if g.SecretKeyID == "" {
			return fmt.Errorf("could not find private key")
		}
		wg.Wait()
	}
	return nil
}
