package internal

import (
	"io"
	"log"
	"net"
	"os"

	"github.com/sirupsen/logrus"
)

const (
	DockerHostEnvVar     = "DOCKER_HOST"
	DockerCertPathEnvVar = "DOCKER_CERT_PATH"

	LocalDockerSocket = "/tmp/docker.sock"
	LocalDockerHost   = "unix://" + LocalDockerSocket

	RemoteDockerSocket = "/var/run/docker.sock"
)

func (r *Remote) Run() error {

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
