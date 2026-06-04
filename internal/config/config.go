package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Db_Url            string `json:"db_url"`
	Current_User_Name string `json:"current_user_name"`
}

func (c Config) SetUser(user string) {
	c.Current_User_Name = user
	url := GetConfigFilePath()
	file, err := os.OpenFile(url, os.O_WRONLY, os.ModeExclusive)
	defer file.Close()
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
	}
	encoder := json.NewEncoder(file)
	fmt.Printf("Filepath: %s\n", url)
	if err = encoder.Encode(c); err != nil {
		fmt.Printf("Error encoding JSON: %v\n", err)
	}
}

func GetConfigFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting config file: %v\n", err)
	}
	url := home + "/.gatorconfig.json"
	return url
}

type State struct {
	ConfigPtr *Config
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	CommandMap map[string]func(*State, Command) error
}

func (c *Commands) Run(s *State, cmd Command) error {
	err := c.CommandMap[cmd.Name](s, cmd)
	return err
}

func (c *Commands) Register(name string, f func(*State, Command) error) {
	c.CommandMap[name] = f
}
