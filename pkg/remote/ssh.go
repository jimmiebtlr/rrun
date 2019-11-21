package remote

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
)

func SshConfig(user string, privateKeyLoc string) (*ssh.ClientConfig, error) {
	file, err := os.Open(privateKeyLoc)
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
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return config, nil
}

//e.g. output, err := remoteRun("root", "MY_IP", "PRIVATE_KEY", "ls")
func Exec(config *ssh.ClientConfig, addr string, workDir string, cmd string, nixConf string) (bytes.Buffer, error) {
	var b bytes.Buffer // import "bytes"

	// Connect
	client, err := ssh.Dial("tcp", net.JoinHostPort(addr, "22"), config)
	if err != nil {
		return b, err
	}
	// Create a session. It is one session per command.
	session, err := client.NewSession()
	if err != nil {
		return b, err
	}
	defer session.Close()

	session.Stderr = os.Stderr // get output
	session.Stdout = &b        // get output
	// you can also pass what gets input to the stdin, allowing you to pipe
	// content from client to server
	//      session.Stdin = bytes.NewBufferString("My input")

	// Finally, run the command
	fullCmd := ". ~/.nix-profile/etc/profile.d/nix.sh && cd " + workDir + " && nix-shell " + nixConf + " --command '" + cmd + "'"
	fmt.Println(fullCmd)
	err = session.Run(fullCmd)
	return b, err
}
