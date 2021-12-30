package internal

import (
	"io/ioutil"
	"log"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
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
		log.Fatal(err)
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
