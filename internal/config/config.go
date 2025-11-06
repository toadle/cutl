package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type FileConfig struct {
	Columns []string `json:"columns"`
}

type Config struct {
	Files map[string]FileConfig `json:"files"`
}

const configFileName = ".cutl_config.json"

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, configFileName), nil
}

func Load() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return &Config{Files: make(map[string]FileConfig)}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Files: make(map[string]FileConfig)}, nil
		}
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return &Config{Files: make(map[string]FileConfig)}, nil
	}

	if config.Files == nil {
		config.Files = make(map[string]FileConfig)
	}

	return &config, nil
}

func (c *Config) Save() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func (c *Config) GetFileConfig(filePath string) (FileConfig, bool) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}
	
	config, exists := c.Files[absPath]
	return config, exists
}

func (c *Config) SetFileConfig(filePath string, fileConfig FileConfig) {
	if c.Files == nil {
		c.Files = make(map[string]FileConfig)
	}
	
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}
	
	c.Files[absPath] = fileConfig
}

func (c *Config) UpdateColumns(filePath string, columns []string) error {
	fileConfig := FileConfig{
		Columns: columns,
	}
	
	c.SetFileConfig(filePath, fileConfig)
	return c.Save()
}