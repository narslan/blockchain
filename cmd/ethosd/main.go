package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"

	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/narslan/ethos/http"
)

// main is the entry point to our application binary. However, it has some poor
// usability so we mainly use it to delegate out to our Main type.
func main() {

	// Setup signal handlers.
	//ctx, _ := context.WithCancel(context.Background())
	ctx := context.Background()
	m := NewMain()

	// Parse command line flags & load configuration.
	if err := m.ParseFlags(ctx, os.Args[1:]); err == flag.ErrHelp {
		os.Exit(1)
	} else if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Execute program.
	if err := m.Run(ctx); err != nil {
		m.Close()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}

// Main represents the program.
type Main struct {
	// Configuration path and parsed config data.
	Config     Config
	ConfigPath string

	// HTTP server for handling HTTP communication.
	// SQLite services are attached to it before running.
	HTTPServer *http.Server
}

// NewMain returns a new instance of Main.
func NewMain() *Main {
	return &Main{
		Config:     DefaultConfig(),
		ConfigPath: DefaultConfigPath,

		HTTPServer: http.NewServer(),
	}
}

// Run executes the program. The configuration should already be set up before
// calling this function.
func (m *Main) Run(ctx context.Context) (err error) {

	m.HTTPServer.Addr = m.Config.HTTP.Addr
	m.HTTPServer.Domain = m.Config.HTTP.Domain
	m.HTTPServer.UseTLS = m.Config.HTTP.UseTLS

	ok := m.HTTPServer.UseTLS
	if ok {
		m.HTTPServer.Cert = m.Config.HTTP.Cert
		m.HTTPServer.Key = m.Config.HTTP.Key
	}

	if err := m.HTTPServer.Open(); err != nil {
		return err
	}

	return nil
}

// Close gracefully stops the program.
func (m *Main) Close() error {
	if m.HTTPServer != nil {
		if err := m.HTTPServer.Close(); err != nil {
			return err
		}
	}

	return nil
}

// ParseFlags parses the command line arguments & loads the config.

func (m *Main) ParseFlags(ctx context.Context, args []string) error {
	// Our flag set is very simple. It only includes a config path.
	fs := flag.NewFlagSet("ethos", flag.ContinueOnError)
	fs.StringVar(&m.ConfigPath, "config", DefaultConfigPath, "config path")
	if err := fs.Parse(args); err != nil {
		return err
	}

	// The expand() function is here to automatically expand "~" to the user's
	// home directory. This is a common task as configuration files are typing
	// under the home directory during local development.
	configPath, err := expand(m.ConfigPath)
	if err != nil {
		return err
	}

	// Read our TOML formatted configuration file.
	config, err := ReadConfigFile(configPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", m.ConfigPath)
	} else if err != nil {
		return err
	}
	m.Config = config

	return nil
}

const (
	// DefaultConfigPath is the default path to the application configuration.
	DefaultConfigPath = "./config.toml"
)

// Config represents the CLI configuration file.
type Config struct {
	HTTP struct {
		Addr   string `toml:"addr"`
		Domain string `toml:"domain"`
		Cert   string `toml:"cert"`
		Key    string `toml:"key"`
		UseTLS bool   `toml:"tls"`
	} `toml:"http"`
}

// DefaultConfig returns a new instance of Config with defaults set.
func DefaultConfig() Config {
	var config Config
	config.HTTP.Addr = ":4443"
	config.HTTP.Domain = "redstar"
	return config
}

// ReadConfigFile unmarshals config from
func ReadConfigFile(filename string) (Config, error) {
	config := DefaultConfig()
	if buf, err := ioutil.ReadFile(filename); err != nil {
		return config, err
	} else if err := toml.Unmarshal(buf, &config); err != nil {
		return config, err
	}
	return config, nil
}

// expand returns path using tilde expansion. This means that a file path that
// begins with the "~" will be expanded to prefix the user's home directory.
func expand(path string) (string, error) {
	// Ignore if path has no leading tilde.
	if path != "~" && !strings.HasPrefix(path, "~"+string(os.PathSeparator)) {
		return path, nil
	}

	// Fetch the current user to determine the home path.
	u, err := user.Current()
	if err != nil {
		return path, err
	} else if u.HomeDir == "" {
		return path, fmt.Errorf("home directory unset")
	}

	if path == "~" {
		return u.HomeDir, nil
	}
	return filepath.Join(u.HomeDir, strings.TrimPrefix(path, "~"+string(os.PathSeparator))), nil
}
