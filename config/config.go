package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

type ConfigFile struct {
	Path    string
	IniFile *ini.File
}

func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	file := filepath.Join(home, "/.aws/credentials")

	return file, nil
}

func LoadConfig(path string) (*ConfigFile, error) {
	config := &ConfigFile{
		Path: path,
	}

	if _, err := os.Stat(path); err == nil {
		if parseErr := config.parseFile(); parseErr != nil {
			return nil, parseErr
		}
	} else {
		return nil, fmt.Errorf("config file %s doesn't exist", path)
	}

	return config, nil
}

func (c *ConfigFile) parseFile() error {
	log.Printf("Parsing config file %s", c.Path)

	f, err := ini.LoadSources(ini.LoadOptions{
		AllowNestedValues:   true,
		InsensitiveSections: false,
		InsensitiveKeys:     true,
	}, c.Path)
	if err != nil {
		return fmt.Errorf("error parsing config file %s: %w", c.Path, err)
	}
	c.IniFile = f

	return nil
}

func (c *ConfigFile) SectionExists(section string) bool {
	_, err := c.IniFile.GetSection(section)

	return err == nil
}

func (c *ConfigFile) SetValue(section string, key string, value string) {
	c.IniFile.Section(section).Key(key).SetValue(value)
}

func (c *ConfigFile) Save() error {
	return c.IniFile.SaveTo(c.Path)
}
