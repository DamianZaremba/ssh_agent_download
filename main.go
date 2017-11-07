// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// ssh_agent_download
// This binary does 2 things;
// * Creates an SSH agent listening on a unix socket
// * Spawns an `ssh` call, pointing the agent process at our unix socket
//
// The aim is to have a very simple way to dump remotly injected keys
//

package main

import (
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh/agent"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"os/user"
	"strconv"
)

func main() {
	usr, _ := user.Current()

	// Get our options
	targetKeyLocation := flag.String("targetKeyLocation", usr.HomeDir+"/.ssh/id_rsa_temporary", "Location to write the SSH key in")
	targetServer := flag.String("targetServer", "localhost", "Target SSH server name")
	targetServerPort := flag.Int("targetServerPort", 22, "Target SSH server port")
	keyLifetime := flag.Int("keyLifetime", 86400, "Seconds the key should exist in the SSH Agent for")
	flag.Parse()

	// Create a temp socket path
	// This is where we will point the SSH child process to
	file, err := ioutil.TempFile(os.TempDir(), "ssh-agent.sock")
	os.Remove(file.Name())       // Remove as we need to bind to it
	defer os.Remove(file.Name()) // Cleanup later

	// Start listening on the unix socket for connections
	s, err := net.Listen("unix", file.Name())
	if err != nil {
		log.Fatal("Could not listen on ssh-agent.sock", err)
	}
	defer s.Close() // Close later

	// Key ring instance
	// This is where all our logic exists for dumping the keys
	// It is called from the crypto/ssh/agent implementation
	keyring := NewExtendedKeyring(*targetKeyLocation)

	// Handling incoming requests (from the SSH process)
	go func() {
		for {
			conn, err := s.Accept()
			if err != nil {
				log.Fatal("Could not accept request", err.Error())
				continue
			}

			// Serve the request in the background
			// SSH can make multiple calls, we don't need to block things
			go func() {
				if err := agent.ServeAgent(keyring, conn); err != nil {
					// Ignore early disconnects from SSH, we can't do anything
					if err != io.EOF {
						log.Fatal("Failed to handle agent request", err.Error())
					}
				}
			}()
		}
	}()

	// Spawn a child process calling ssh
	sshCmd := exec.Command("ssh", "-p", strconv.Itoa(*targetServerPort), *targetServer)
	// Ensure the SSH agent points to us :)
	sshCmd.Env = []string{fmt.Sprintf("SSH_AUTH_SOCK=%s", file.Name())}
	// Map std{in,out,err} to the parent process (us)
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	// Execute the command
	// At this point, this looks like normal SSH to the user
	fmt.Printf("ssh_agent_download: connecting to %s\n\n", *targetServer)
	sshCmd.Run()

	// Once the SSH command has finished
	// Run `ssh-add` against the key we should have written out
	// This ensures the key 'magically' exists in the currect working environment
	fmt.Printf("\nssh_agent_download: adding %s to keychain\n", *targetKeyLocation)
	addCmd := exec.Command("ssh-add", "-t", strconv.Itoa(*keyLifetime), *targetKeyLocation)
	addCmd.Run()

	// And done :)
}
