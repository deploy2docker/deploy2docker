package internal

import (
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type SSHClient struct {
	address string
	user    string
	timeout time.Duration // default 30 seconds
}

func NewSSHClient(address string, user string) *SSHClient {
	return &SSHClient{
		address: address,
		user:    user,
		timeout: 30 * time.Second,
	}
}

func (sc *SSHClient) ConnectWithPassword(password string) (*ssh.Client, error) {
	config := ssh.ClientConfig{
		User: sc.user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		Timeout:         sc.timeout,
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
	}

	return sc.Dial(config)
}

func (sc *SSHClient) ConnectWithKey(key string) (*ssh.Client, error) {
	pemBytes, err := ioutil.ReadFile(key)
	if err != nil {
		log.Fatal(err)
	}
	signer, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}
	config := ssh.ClientConfig{
		User: sc.user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		Timeout:         sc.timeout,
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
	}

	return sc.Dial(config)
}

func (sc *SSHClient) Dial(config ssh.ClientConfig) (*ssh.Client, error) {
	conn, err := ssh.Dial("tcp", sc.address, &config)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
