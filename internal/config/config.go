package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Connection struct {
	Name        string `toml:"name"`
	SSHHost     string `toml:"ssh_host"`
	SSHUser     string `toml:"ssh_user"`
	SSHPort     int    `toml:"ssh_port,omitempty"`
	SSHKeyPath  string `toml:"ssh_key,omitempty"`
	SSHPassword string `toml:"ssh_password,omitempty"`
	DBHost      string `toml:"db_host,omitempty"`
	DBPort      int    `toml:"db_port,omitempty"`
	DBName      string `toml:"db_name"`
	DBUser      string `toml:"db_user"`
	DBPassword  string `toml:"db_password,omitempty"`
}

type Config struct {
	Connections []Connection `toml:"connections"`
}

func Dir() string {
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		return filepath.Join(d, "dioptra")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "dioptra")
}

func Path() string {
	return filepath.Join(Dir(), "config.toml")
}

func Load() (*Config, error) {
	var cfg Config
	_, err := toml.DecodeFile(Path(), &cfg)
	if os.IsNotExist(err) {
		return &cfg, nil
	}
	return &cfg, err
}

func (c *Config) Save() error {
	dir := Dir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	f, err := os.OpenFile(Path(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(c)
}

func (c *Config) Get(name string) (*Connection, error) {
	for i := range c.Connections {
		if c.Connections[i].Name == name {
			return &c.Connections[i], nil
		}
	}
	return nil, fmt.Errorf("connection %q not found", name)
}

func (c *Config) Add(conn Connection) error {
	for _, existing := range c.Connections {
		if existing.Name == conn.Name {
			return fmt.Errorf("connection %q already exists", conn.Name)
		}
	}
	c.Connections = append(c.Connections, conn)
	return c.Save()
}

func (c *Config) Remove(name string) error {
	for i, conn := range c.Connections {
		if conn.Name == name {
			c.Connections = append(c.Connections[:i], c.Connections[i+1:]...)
			return c.Save()
		}
	}
	return fmt.Errorf("connection %q not found", name)
}
