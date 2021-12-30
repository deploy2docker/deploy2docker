package internal

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
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
		Address  *string `json:"address,omitempty"`
		User     *string `json:"user,omitempty"`
		Key      *string `json:"key,omitempty"`
		Password *string `json:"password,omitempty"`
	}
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Load(path string) error {
	return json.Unmarshal([]byte(path), c)
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

	c.Remote.Address = &result
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

	c.Remote.User = &result
	return nil
}

func (c *Config) PromptRemoteKey() error {
	validate := func(input string) error {
		if input == "" {
			return errors.New("remote key cannot be empty")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Remote key",
		Validate: validate,
	}

	result, err := prompt.Run()
	if err != nil {
		return err
	}

	c.Remote.Key = &result
	return nil
}

func (c *Config) PromptRemotePassword() error {
	validate := func(input string) error {
		if input == "" {
			return errors.New("remote password cannot be empty")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Remote password",
		Validate: validate,
	}

	result, err := prompt.Run()
	if err != nil {
		return err
	}

	c.Remote.Password = &result
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

	err = c.PromptRemoteKey()
	if err != nil {
		return err
	}

	err = c.PromptRemotePassword()
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
