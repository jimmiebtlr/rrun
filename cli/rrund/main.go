package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"rrun/pkg/remote"

	"github.com/BurntSushi/toml"
	"golang.org/x/crypto/ssh"
)

const (
	ConfigFilename    = "rrun.toml"
	NixConfigFilename = "dev-env.nix"
)

type RunCmd struct {
	RunDir  string `json:"runDir"`
	Command string `json:"command"`
}

type Config struct {
	RootDir          string
	RemoteRoot       string
	RemoteWorkingDir string

	// This will go away hopefully when remote provisioning is implemented
	Ssh struct {
		KeyfileLoc string `toml:"keyfile"`
		RemoteUser string `toml:"user"`
		RemoteAddr string `toml:"addr"`
		RemotePort string `toml:"port"`
	}

	Tunnels []struct {
		RemotePort string `toml:"remotePort"`
		LocalPort  string `toml:"localPort"`
	} `toml:"tunnels"`
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
var activeConnections = map[string]*ActiveConnection{}

func Start(cfg Config) {
	sshCfg, err := cfg.SshConfig()
	if err != nil {
		panic(err)
	}

	// Store in ActiveConnections
	activeConnections[cfg.RootDir] = &ActiveConnection{Config: cfg}

	// Start sync, use full rootdir to place it in ~/cfg.RootDir location (so it's in correct permissioned place)
	go remote.Sync(cfg.Ssh.RemoteUser, cfg.Ssh.RemoteAddr, cfg.RootDir, cfg.RemoteRoot)

	// Start tunnels
	for _, t := range cfg.Tunnels {
		fmt.Println("Starting tunnel ", t.RemotePort, t.LocalPort)
		go remote.Tunnel(sshCfg, cfg.Ssh.RemoteAddr, t.RemotePort, t.LocalPort)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	req := RunCmd{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println(err)
		return
	}

	nearestParent, err := NearestParent(req.RunDir, ConfigFilename)
	if err != nil {
		panic(err)
	}

	nixConf, err := NearestParent(req.RunDir, NixConfigFilename)
	if err != nil {
		panic(err)
	}

	ac := activeConnections[nearestParent]
	if ac == nil {
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

		conf.RootDir = nearestParent
		conf.RemoteRoot = "~/" + strings.Replace(nearestParent, "/", ".", -1)
		rd, _ := filepath.Abs(req.RunDir)
		conf.RemoteWorkingDir = filepath.Join(conf.RemoteRoot, strings.Replace(rd, nearestParent, "", -1))

		Start(conf)

		ac = &ActiveConnection{Config: conf}
	}

	cfg, err := ac.Config.SshConfig()
	if err != nil {
		panic(err)
	}

	b, err := remote.Exec(cfg, ac.Config.Ssh.RemoteAddr, ac.Config.RemoteWorkingDir, req.Command, ac.Config.RemoteRoot+strings.Replace(filepath.Join(nixConf, NixConfigFilename), nearestParent, "", -1))
	if err != nil {
		fmt.Println("ERROR")
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	b.WriteTo(w)
}

func NearestParent(startingDir string, file string) (parentDir string, err error) {
	dir, err := filepath.Abs(startingDir)
	if err != nil {
		return "", err
	}

	for {
		fmt.Println("Looking in " + dir)

		if stat, err := os.Stat(path.Join(dir, file)); err == nil && stat != nil {
			return dir, nil
		}

		if dir == path.Dir(dir) {
			return "", errors.New("Config file not found in any parent dirs")
		}

		dir = path.Dir(dir)
	}
}

func main() {
	fmt.Println("Starting rrun daemon!")
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":5059", nil))
}
