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
	"errors"
	"os"
	"time"

	"github.com/deploy2docker/deploy2docker/internal"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "Deploy To Docker",
		Usage: "Deploy to Remote Docker using SSH",
		Commands: []*cli.Command{
			{
				Name:    "init",
				Aliases: []string{"i"},
				Usage:   "Initialize a new configuration file",
				Action: func(c *cli.Context) error {
					config := internal.NewConfig()
					return config.Init()
				},
			},
			{
				Name:    "deploy",
				Aliases: []string{"d"},
				Usage:   "Deploy to remote docker",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "debug",
						Aliases: []string{"d"},
						Usage:   "Enable debug mode",
						Value:   false,
						EnvVars: []string{"DEBUG"},
					},
					&cli.PathFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Usage:   "Configuration file",
						EnvVars: []string{"CONFIG"},
						Value:   "deploy2docker.json",
					},
					&cli.PathFlag{
						Name:     "private-key",
						Aliases:  []string{"k"},
						Usage:    "Private key file",
						Value:    "",
						EnvVars:  []string{"PRIVATE_KEY"},
						Required: false,
					},
					&cli.StringFlag{
						Name:     "password",
						Aliases:  []string{"p"},
						Usage:    "Password",
						Value:    "",
						EnvVars:  []string{"SSH_PASSWORD"},
						Required: false,
					},
					&cli.PathFlag{
						Name:  "path",
						Usage: "Path to the directory containing the dockerfile.",
						Value: "",
					},
				},
				Action: func(c *cli.Context) error {

					if c.Bool("debug") {
						logrus.SetLevel(logrus.DebugLevel)
					}

					if c.String("config") == "" {
						return errors.New("configuration file is required")
					}

					// check if config file exists
					if _, err := os.Stat(c.String("config")); os.IsNotExist(err) {
						return errors.New("configuration file does not exist")
					}

					logrus.Debugln("Reading configuration file", c.String("config"))

					config := internal.NewConfig()
					err := config.Load(c.String("config"))
					if err != nil {
						return err
					}

					err = config.Validate()
					if err != nil {
						return err
					}

					remote := internal.NewRemote(internal.RemoteConfig{
						Address: config.Remote.Address,
						User:    config.Remote.User,
						Timeout: time.Second * 10,
					})

					if c.String("private-key") != "" {
						err := remote.ConnectWithKey(c.String("private-key"))
						if err != nil {
							logrus.Debug(err)
							return err
						}
					} else if c.String("password") != "" {
						err := remote.ConnectWithPassword(c.String("password"))
						if err != nil {
							return err
						}
					} else {
						logrus.Warnln("No key or password provided")
					}

					err = remote.Connect()
					if err != nil {
						return err
					}

					docker, err := internal.NewDocker()
					if err != nil {
						return err
					}

					logrus.Debugln("Connected to remote server")

					err = docker.Close()
					if err != nil {
						return err
					}

					// current working directory
					cwd, err := os.Getwd()
					if err != nil {
						return err
					}

					if c.String("path") != "" {
						cwd = c.String("path")
					}

					logrus.Infoln("Building image")
					err = docker.Build(c.Context, cwd, []string{
						config.Service.Name,
					})
					if err != nil {
						return err
					}

					logrus.Infoln("Built successfully image ", config.Service.Name)

					err = docker.Run(c.Context, config)
					if err != nil {
						return err
					}

					logrus.Infoln("Service ", config.Service.Name, " is running")

					err = remote.Close()
					if err != nil {
						return err
					}

					logrus.Debugln("Closed connection to remote server")
					return nil

				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		logrus.Error(err)
	}
}
