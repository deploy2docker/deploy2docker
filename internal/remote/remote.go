/*
MIT License

Copyright (c) 2021 Deploy to Docker (satish.babariya@gmail.com)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NON INFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package remote

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

const (
	DockerHostEnvVar     = "DOCKER_HOST"
	DockerCertPathEnvVar = "DOCKER_CERT_PATH"
	LocalDockerSocket    = "/tmp/docker.sock"
	LocalDockerHost      = "unix://" + LocalDockerSocket
	RemoteDockerSocket   = "/var/run/docker.sock"
)

type Remote struct {
	client  *ssh.Client
	config  *ssh.ClientConfig
	address string
}

type RemoteConfig struct {
	Address string
	User    string
	Timeout time.Duration
}

func ParseRemoteConfig(address string) (*RemoteConfig, error) {
	cfg := &RemoteConfig{
		Timeout: time.Second * 10,
	}

	// loop through the address and split it into user and host
	for i := len(address) - 1; i >= 0; i-- {
		if address[i] == '@' {
			cfg.User = address[:i]
			cfg.Address = address[i+1:]
			break
		}
	}

	if cfg.Address == "" {
		return nil, fmt.Errorf("invalid address: %s", address)
	}

	hasPort := false
	for i := len(cfg.Address) - 1; i >= 0; i-- {
		if cfg.Address[i] == ':' {
			hasPort = true
			break
		}
	}

	if !hasPort {
		cfg.Address = fmt.Sprint(cfg.Address, ":22")
	}

	return cfg, nil
}

func NewRemote(config RemoteConfig) *Remote {
	return &Remote{
		address: config.Address,
		config: &ssh.ClientConfig{
			User:            config.User,
			Timeout:         config.Timeout,
			HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
		},
	}
}

func (r *Remote) ConnectWithPassword(password string) error {
	r.config.Auth = []ssh.AuthMethod{
		ssh.Password(password),
	}

	conn, err := r.dial()
	if err != nil {
		return err
	}

	r.client = conn
	return nil
}

func (r *Remote) ConnectWithKey(key string) error {
	pemBytes, err := ioutil.ReadFile(key)
	if err != nil {
		return err
	}
	signer, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		return err
	}
	r.config.Auth = []ssh.AuthMethod{
		ssh.PublicKeys(signer),
	}

	conn, err := r.dial()
	if err != nil {
		return err
	}

	r.client = conn
	return nil
}

func (r *Remote) Connect() error {
	if r.client != nil {
		return nil
	}

	conn, err := r.dial()
	if err != nil {
		return err
	}

	r.client = conn
	return nil
}

func (r *Remote) dial() (*ssh.Client, error) {
	conn, err := ssh.Dial("tcp", r.address, r.config)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (r *Remote) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	if err := os.Remove(LocalDockerHost); err != nil {
		return err
	}
	return nil
}

func (r *Remote) PorxyDockerSocket() error {

	// remove the docker socket if it exists
	err := r.Cleanup()
	if err != nil {
		return err
	}

	// set docker host env var
	// os.Setenv("DOCKER_HOST", LocalDockerHost)

	// establish connection with remote docker
	remote, err := r.client.Dial("unix", RemoteDockerSocket)
	if err != nil {
		return fmt.Errorf("failed to connect to remote docker: %s", err)
	}
	// defer remote.Close()

	// start the local docker socket
	local, err := net.Listen("unix", LocalDockerSocket)
	if err != nil {
		return fmt.Errorf("failed to start local docker socket: %s", err)
	}
	// defer local.Close()

	// forward the connection between the two sockets
	go func() {
		for {
			client, err := local.Accept()
			if err != nil {
				logrus.Errorln(err)
				return
			}

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
	}()

	return nil
}

func (r *Remote) Cleanup() error {
	if _, err := os.Stat(LocalDockerSocket); err == nil {
		if err := os.RemoveAll(LocalDockerSocket); err != nil {
			return err
		}
	}
	return nil
}
