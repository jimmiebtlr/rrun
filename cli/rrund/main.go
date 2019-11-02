package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/davecgh/go-spew/spew"
	"github.com/jimmiebtlr/rdev/pkg/remote"
	"golang.org/x/crypto/ssh"
)

const (
	ConfigFilename = ".rdev.toml"
)

type RunCmd struct {
	RunDir  string
	Command string
}

type Config struct {
	RootDir string

	// This will go away hopefully when remote provisioning is implemented
	Ssh struct {
		KeyfileLoc string `toml:"keyfile"`
		RemoteUser string `toml:"user"`
		RemoteAddr string `toml:"addr"`
		RemotePort string `toml:"port"`
	}

	Tunnels []struct {
		RemotePort string
		LocalPort  string
	}
}

func (ac *Config) SshConfig() (*ssh.ClientConfig, error) {
	file, err := os.Open(ac.Ssh.KeyfileLoc)
	privateKey, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// privateKey could be read from a file, or retrieved from another storage
	// source, such as the Secret Service / GNOME Keyring
	key, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		fmt.Println("Failed parsing")
		return nil, err
	}

	// Authentication
	config := &ssh.ClientConfig{
		User: ac.Ssh.RemoteUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return config, nil
}

type ActiveConnection struct {
	Config Config
}

// map key is RootDir
var ActiveConnections = map[string]ActiveConnection{}

func Start(cfg Config) {
	sshCfg, err := cfg.SshConfig()
	if err != nil {
		panic(err)
	}

	// Store in ActiveConnections
	ActiveConnections[cfg.RootDir] = ActiveConnection{Config: cfg}

	// Start sync, use full rootdir to place it in ~/cfg.RootDir location (so it's in correct permissioned place)
	go remote.Sync(cfg.Ssh.RemoteAddr, cfg.RootDir, "."+cfg.RootDir)

	// Start tunnels
	for _, t := range cfg.Tunnels {
		go remote.Tunnel(sshCfg, cfg.Ssh.RemoteAddr, t.RemotePort, t.LocalPort)
	}
}

func handleRun() {
	// Find parent with config
	// Check if in activeConnections
	// if not Parse Toml and start
	// Run command
}

func NearestParent(startingDir string) (parentDir string, err error) {
	dir, err := filepath.Abs(startingDir)
	if err != nil {
		return "", err
	}

	for {
		fmt.Println("Looking in " + dir)

		if stat, err := os.Stat(path.Join(dir, ConfigFilename)); err == nil && stat != nil {
			return dir, nil
		}

		if dir == path.Dir(dir) {
			return "", errors.New("Config file not found in any parent dirs")
		}

		dir = path.Dir(dir)
	}
}

func main() {
	nearestParent, err := NearestParent("./ai_architect")
	if err != nil {
		panic(err)
	}

	cfgFile, err := os.Open(filepath.Join(nearestParent, ConfigFilename))
	if err != nil {
		panic(err)
	}

	tomlData, err := ioutil.ReadAll(cfgFile)
	if err != nil {
		panic(err)
	}

	var conf Config
	if _, err := toml.Decode(string(tomlData), &conf); err != nil {
		// handle error
		panic(err)
	}

	spew.Dump(conf)
}
