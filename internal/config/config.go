package config

import (
	"encoding/json"
	"os"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}



func getConfigFilePath() (string, error) {
	homepath, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := homepath + "/" + configFileName
	return path, nil
}

func write(cfg Config) error {
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}
	Val, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	err = os.WriteFile(path, Val, 0644)
	if err != nil {
		return err
	}
	return nil
}

func Read() (Config, error) {
	var Val Config
	path, err := getConfigFilePath()
	if err != nil {
		return Val, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return Val, err
	}

	if err = json.Unmarshal(content, &Val); err != nil {
		return Val, err
	}

	return Val, nil
}

func (c *Config) SetUser(user string) error {
	c.CurrentUserName = user
	err := write(*c)
	if err != nil {
		return err
	}
	return nil
}


