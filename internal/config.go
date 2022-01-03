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

package internal

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
)

type Config struct {

	// Service
	Service struct {
		Name       string `json:"name"`
		Ports      []int  `json:"ports,omitempty"`
		Dockerfile string `json:"dockerfile"`
	}

	// Remote
	Remote struct {
		Address string `json:"address,omitempty"`
		User    string `json:"username,omitempty"`
	}
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Load(path string) error {
	logrus.Debugln("Loading config from", path)
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, c)
	if err != nil {
		return err
	}
	return nil

}

func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0644)
}

func (c *Config) Init() error {
	err := c.PromptServiceName()
	if err != nil {
		return err
	}

	err = c.PromptServicePorts()
	if err != nil {
		return err
	}

	err = c.PromptServiceDockerfile()
	if err != nil {
		return err
	}

	err = c.PromptRemote()
	if err != nil {
		return err
	}

	return c.Save("deploy2docker.json")
}

func (c *Config) PromptServiceName() error {
	validate := func(input string) error {
		if input == "" {
			return errors.New("service name cannot be empty")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Service name",
		Validate: validate,
	}

	result, err := prompt.Run()
	if err != nil {
		return err
	}

	c.Service.Name = result
	return nil
}

func (c *Config) PromptRemoteAddress() error {
	validate := func(input string) error {
		if input == "" {
			return errors.New("remote address cannot be empty")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Remote address",
		Validate: validate,
	}

	result, err := prompt.Run()
	if err != nil {
		return err
	}

	c.Remote.Address = result
	return nil
}

func (c *Config) PromptRemoteUser() error {
	validate := func(input string) error {
		if input == "" {
			return errors.New("remote user cannot be empty")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Remote user",
		Validate: validate,
	}

	result, err := prompt.Run()
	if err != nil {
		return err
	}

	c.Remote.User = result
	return nil
}

func (c *Config) PromptServicePorts() error {
	validate := func(input string) error {
		if input == "" {
			return errors.New("service ports cannot be empty")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Service ports",
		Validate: validate,
	}

	result, err := prompt.Run()
	if err != nil {
		return err
	}

	c.Service.Ports = []int{}
	for _, port := range strings.Split(result, ",") {
		port = strings.TrimSpace(port)
		if port == "" {
			continue
		}

		p, err := strconv.Atoi(port)
		if err != nil {
			return err
		}

		c.Service.Ports = append(c.Service.Ports, p)
	}

	return nil
}

func (c *Config) PromptServiceDockerfile() error {
	validate := func(input string) error {
		if input == "" {
			return errors.New("service dockerfile cannot be empty")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Service dockerfile",
		Validate: validate,
	}

	result, err := prompt.Run()
	if err != nil {
		return err
	}

	c.Service.Dockerfile = result
	return nil
}

func (c *Config) PromptRemote() error {
	err := c.PromptRemoteAddress()
	if err != nil {
		return err
	}

	err = c.PromptRemoteUser()
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) PromptService() error {
	err := c.PromptServiceName()
	if err != nil {
		return err
	}

	err = c.PromptServicePorts()
	if err != nil {
		return err
	}

	err = c.PromptServiceDockerfile()
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) Validate() error {
	if c.Service.Name == "" {
		return errors.New("service name cannot be empty")
	}

	if c.Service.Dockerfile == "" {
		return errors.New("service dockerfile cannot be empty")
	}

	if len(c.Service.Ports) == 0 {
		return errors.New("service ports cannot be empty")
	}

	if c.Remote.Address == "" {
		return errors.New("remote address cannot be empty")
	}

	if c.Remote.User == "" {
		return errors.New("remote user cannot be empty")
	}

	return nil
}
