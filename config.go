package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
	"gopkg.in/yaml.v2"
)

type Config struct {
	OutputDir      string   `yaml:"output_dir"`
	Concurrency    int      `yaml:"concurrency"`
	RequestTimeout string   `yaml:"request_timeout"`
	RetryDelay     string   `yaml:"retry_delay"`
	MaxRetries     int      `yaml:"max_retries"`
	Prefixes       []string `yaml:"prefixes"`
	APIKeys        []string `yaml:"api_keys"`
}

func DefaultConfig() *Config {
	return &Config{
		OutputDir:      ".",
		Concurrency:    20,
		RequestTimeout: "20s",
		RetryDelay:     "50ms",
		MaxRetries:     3,
		Prefixes:       []string{},
		APIKeys:        []string{},
	}
}

func LoadConfig() *Config {
	config := DefaultConfig()

	if data, err := ioutil.ReadFile("config.yaml"); err == nil {
		var yamlConfig Config
		if err := yaml.Unmarshal(data, &yamlConfig); err == nil {
			if yamlConfig.OutputDir != "" {
				config.OutputDir = yamlConfig.OutputDir
			}
			if yamlConfig.Concurrency > 0 {
				config.Concurrency = yamlConfig.Concurrency
			}
			if yamlConfig.RequestTimeout != "" {
				config.RequestTimeout = yamlConfig.RequestTimeout
			}
			if yamlConfig.RetryDelay != "" {
				config.RetryDelay = yamlConfig.RetryDelay
			}
			if yamlConfig.MaxRetries > 0 {
				config.MaxRetries = yamlConfig.MaxRetries
			}
			if len(yamlConfig.Prefixes) > 0 {
				config.Prefixes = yamlConfig.Prefixes
			}
			if len(yamlConfig.APIKeys) > 0 {
				config.APIKeys = yamlConfig.APIKeys
			}
		}
	}

	if outputDir := os.Getenv("OUTPUT_DIR"); outputDir != "" {
		config.OutputDir = outputDir
	}

	return config
}

func (c *Config) GetRequestTimeout() time.Duration {
	duration, err := time.ParseDuration(c.RequestTimeout)
	if err != nil {
		return 20 * time.Second 
	}
	return duration
}

func (c *Config) GetRetryDelay() time.Duration {
	duration, err := time.ParseDuration(c.RetryDelay)
	if err != nil {
		return 50 * time.Millisecond 
	}
	return duration
}

func (c *Config) GetOutputFileName() string {
	timestamp := time.Now().Format("20060102")
	filename := "report_" + timestamp + ".html"
	return filepath.Join(c.OutputDir, filename)
}
