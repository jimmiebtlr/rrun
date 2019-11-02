package remote

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
)

// Get default location of a private key
func privateKeyPath() string {
	return os.Getenv("HOME") + "/.ssh/id_rsa"
}

// Get private key for ssh authentication
func parsePrivateKey(keyPath string) (ssh.Signer, error) {
	buff, _ := ioutil.ReadFile(keyPath)
	return ssh.ParsePrivateKey(buff)
}

// Handle local client connections and tunnel data to the remote serverq
// Will use io.Copy - http://golang.org/pkg/io/#Copy
func handleClient(client net.Conn, remote net.Conn) {
	defer client.Close()
	chDone := make(chan bool)

	// Start remote -> local data transfer
	go func() {
		_, err := io.Copy(client, remote)
		if err != nil {
			log.Println("error while copy remote->local:", err)
		}
		chDone <- true
	}()

	// Start local -> remote data transfer
	go func() {
		_, err := io.Copy(remote, client)
		if err != nil {
			log.Println(err)
		}
		chDone <- true
	}()

	<-chDone
}

func Tunnel(clientConfig *ssh.ClientConfig, addr string, remotePort string, localPort string) {
	localAddr := "localhost:" + localPort
	remoteAddr := "localhost:" + remotePort

	// Establish connection with SSH server
	conn, err := ssh.Dial("tcp", net.JoinHostPort(addr, "22"), clientConfig)
	if err != nil {
		fmt.Println("ERROR CONN")
		log.Fatalln(err)
	}
	defer conn.Close()

	// Establish connection with remote server
	remote, err := net.Dial("tcp", localAddr)
	if err != nil {
		log.Fatalln(err)
	}

	// Start local server to forward traffic to remote connection
	local, err := conn.Listen("tcp", remoteAddr)
	if err != nil {
		log.Fatalln(err)
	}
	defer local.Close()

	// Handle incoming connections
	for {
		client, err := local.Accept()
		if err != nil {
			log.Fatalln(err)
		}

		handleClient(client, remote)
	}
}
