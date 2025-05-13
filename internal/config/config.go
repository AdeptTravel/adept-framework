package config

import (
    "os"
    "gopkg.in/yaml.v3"
)

// Config mirrors config.yaml.
type Config struct {
    App struct {
        UseDevHost bool   `yaml:"use_dev_host"`
        DevHost    string `yaml:"dev_host"`
    } `yaml:"app"`
}

// Load reads and unmarshals the YAML config file.
func Load(path string) (Config, error) {
    raw, err := os.ReadFile(path)
    if err != nil {
        return Config{}, err
    }
    var cfg Config
    if err := yaml.Unmarshal(raw, &cfg); err != nil {
        return Config{}, err
    }
    return cfg, nil
}
