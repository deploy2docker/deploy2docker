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
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package main

import (
	"context"
	"os"

	"github.com/deploy2docker/deploy2docker/internal/config"
	"github.com/deploy2docker/deploy2docker/internal/docker"
	"github.com/deploy2docker/deploy2docker/internal/remote"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	var (
		remoteAddress string
		configPath    string
		password      string
		keyPath       string
	)

	app := &cli.App{}
	app.Name = "deploy2docker"
	app.Usage = "Deploy to Docker"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			Usage:       "Path to config file",
			Destination: &configPath,
			TakesFile:   true,
			Value:       "deploy2docker.yaml",
			EnvVars:     []string{"DEPLOY2DOCKER_CONFIG"},
		},
		&cli.StringFlag{
			Name:        "remote",
			Usage:       "Remote address.",
			Required:    true,
			EnvVars:     []string{"DEPLOY2DOCKER_REMOTE"},
			Destination: &remoteAddress,
		},
		&cli.StringFlag{
			Name:  "password",
			Usage: "Password for the remote host.",
			EnvVars: []string{
				"DEPLOY2DOCKER_PASSWORD",
			},
			Destination: &password,
		},
		&cli.StringFlag{
			Name:  "key",
			Usage: "Path to the private key for the remote host.",
			EnvVars: []string{
				"DEPLOY2DOCKER_KEY",
			},
			Destination: &keyPath,
			TakesFile:   true,
		},
	}

	app.Action = func(c *cli.Context) error {
		config, err := config.Parse(configPath)
		if err != nil {
			return err
		}

		if err := config.Validate(); err != nil {
			return err
		}

		r, err := remote.ParseRemoteConfig(remoteAddress)
		if err != nil {
			return err
		}

		// ssh to remote docker host
		remote := remote.NewRemote(*r)

		if password != "" {
			err = remote.ConnectWithPassword(password)
			if err != nil {
				return err
			}
		} else if keyPath != "" {
			err = remote.ConnectWithKey(keyPath)
			if err != nil {
				return err
			}
		} else {
			err = remote.Connect()
			if err != nil {
				return err
			}
		}

		err = remote.PorxyDockerSocket()
		if err != nil {
			return err
		}

		docker, err := docker.NewDockerClient()
		if err != nil {
			return err
		}

		if docker.Ping(context.Background()) {
			println("Docker is running")
		} else {
			println("Docker is not running")
		}

		defer remote.Close()
		defer docker.Close()

		return nil
	}

	// Run the app.
	if err := app.Run(os.Args); err != nil {
		// Log the error and exit.
		logrus.Errorln(err)
	}
}
