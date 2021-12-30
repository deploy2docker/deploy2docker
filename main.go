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
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "Enable debug mode",
				Value:   false,
				EnvVars: []string{"DEBUG"},
			},
			&cli.StringFlag{
				Name:    "address",
				Aliases: []string{"a"},
				Usage:   "Remote SSH address",
				Value:   "",
				EnvVars: []string{"ADDRESS"},
			},
			&cli.StringFlag{
				Name:    "user",
				Aliases: []string{"u"},
				Usage:   "Remote SSH user",
				Value:   "",
				EnvVars: []string{"USER"},
			},
			&cli.StringFlag{
				Name:    "key",
				Aliases: []string{"k"},
				Usage:   "Private key file",
				Value:   "",
				EnvVars: []string{"KEY"},
			},
			&cli.StringFlag{
				Name:    "password",
				Aliases: []string{"p"},
				Usage:   "Password",
				Value:   "",
				EnvVars: []string{"PASSWORD"},
			},
		},
		Action: func(c *cli.Context) error {

			if c.Bool("debug") {
				logrus.SetLevel(logrus.DebugLevel)
			}

			remote := internal.NewRemote(internal.RemoteConfig{
				Address: c.String("address"),
				User:    c.String("user"),
				Timeout: time.Second * 10,
			})

			if c.String("key") != "" {
				err := remote.ConnectWithKey(c.String("key"))
				if err != nil {
					return err
				}
			} else if c.String("password") != "" {
				err := remote.ConnectWithPassword(c.String("password"))
				if err != nil {
					return err
				}
			} else {
				logrus.Warnln("No key or password provided")
				return nil
			}

			err := remote.Connect()
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

			err = remote.Close()
			if err != nil {
				return err
			}

			logrus.Debugln("Closed connection to remote server")
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		logrus.Error(err)
	}
}
