// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"log"
	"os"
	"sync"
	"time"
)

// This is what we're interested in
func (r *extendedKeyring) Add(key agent.AddedKey) error {
	// Figure out what key type we're trying to inject
	switch key.PrivateKey.(type) {
	case *rsa.PrivateKey:
		// Load the injected key
		data := &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key.PrivateKey.(*rsa.PrivateKey)),
		}

		// Ensure the key doesn't exist
		os.Remove(r.targetKeyLocation)

		// Write the key out to disk
		var file, err = os.Create(r.targetKeyLocation)
		if err != nil {
			log.Fatal("\nssh_agent_download: Could not create key file", err.Error())
		}
		defer file.Close() // Ensure we close the file later

		// Secure before writing
		// Note: Technically someone could write in here before we do this
		os.Chmod(r.targetKeyLocation, 0400)

		// Dump the (un-encrypted) key into this file
		pem.Encode(file, data)

		// Let the keyboard monkey know
		fmt.Printf("ssh_agent_download: saved key to %s\n", r.targetKeyLocation)

	// Let the user know this won't work
	default:
		log.Fatal("ssh_agent_download: unsupported key type %T", key.PrivateKey)
	}

	return nil
}

// Our extended entrypoint
func NewExtendedKeyring(targetKeyLocation string) agent.Agent {
	k := &extendedKeyring{}
	k.targetKeyLocation = targetKeyLocation
	return k
}

type extendedKeyring struct {
	keyring
	targetKeyLocation string
}

//
// Upstream code
// See https://github.com/golang/crypto/blob/master/ssh/agent/keyring.go
//

// Internal Types
type privKey struct {
	signer  ssh.Signer
	comment string
	expire  *time.Time
}

type keyring struct {
	mu   sync.Mutex
	keys []privKey

	locked     bool
	passphrase []byte
}

// Upstream interface methods
func (r *extendedKeyring) RemoveAll() error {
	return nil
}

func (r *extendedKeyring) Remove(key ssh.PublicKey) error {
	return nil
}

func (r *extendedKeyring) Unlock(passphrase []byte) error {
	return nil
}

func (r *extendedKeyring) Lock(passphrase []byte) error {
	return nil
}

func (r *extendedKeyring) List() ([]*agent.Key, error) {
	return nil, nil
}

func (r *extendedKeyring) Sign(key ssh.PublicKey, data []byte) (*ssh.Signature, error) {
	return nil, nil
}

func (r *extendedKeyring) Signers() ([]ssh.Signer, error) {
	return nil, nil
}
