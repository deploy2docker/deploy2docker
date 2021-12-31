package internal

import (
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
	return nil
}

func (r *Remote) Connect() error {

	// remove the docker socket if it exists
	err := r.Cleanup()
	if err != nil {
		return err
	}

	// set docker host env var
	os.Setenv("DOCKER_HOST", LocalDockerHost)

	// establish connection with remote docker
	remote, err := r.client.Dial("unix", RemoteDockerSocket)
	if err != nil {
		return err
	}
	// defer remote.Close()

	// start the local docker socket
	local, err := net.Listen("unix", LocalDockerSocket)
	if err != nil {
		return err
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
