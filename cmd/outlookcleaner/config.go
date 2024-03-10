package main

import (
	"os"

	"github.com/raokrutarth/golang-playspace/pkg/logger"
	"github.com/spf13/viper"
)

type (
	// Config stores complete configuration
	Config struct {
		Database DatabaseConfig   `mapstructure:"db"`
		Mail     MailConfig       `mapstructure:"mail"`
		Encrypt  EncryptionConfig `mapstructure:"auth-cli"`
	}

	EncryptionConfig struct {
		Secret string `mapstructure:"secret"`
		Iv     string `mapstructure:"iv"`
		Salt   string `mapstructure:"salt"`
	}

	DatabaseConfig struct {
		Hostname string `mapstructure:"host"`
		Port     int    `mapstructure:"port" default:"5432"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Database string `mapstructure:"database"`
	}

	MailConfig struct {
		Accounts []MailAccountConfig `mapstructure:"accounts"`
	}

	MailAccountConfig struct {
		Hostname    string              `mapstructure:"host"`
		Port        int                 `mapstructure:"port"`
		EncUser     string              `mapstructure:"user"`
		EncPassword string              `mapstructure:"password"`
		Prune       MailboxActionConfig `mapstructure:"prune"`
		Ingest      MailboxActionConfig `mapstructure:"ingest"`
	}
	MailboxActionConfig struct {
		ThresholdDays int      `mapstructure:"threshold_days,omitempty"`
		Folders       []string `mapstructure:"folders"`
	}
)

// getConfig returns the application configuration and secrets
func getConfig() Config {
	log := logger.GetLogger()

	cfg := viper.New()
	cfg.AddConfigPath("./")
	cfg.AddConfigPath("/workspaces/golang-playspace/cmd/outlookcleaner")
	cfg.SetConfigType("yaml")
	cfg.SetConfigName(".secrets")
	err := cfg.ReadInConfig()
	if err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			contextDir, _ := os.Getwd()
			log.Error("unable to find any configuration files", "cwd", contextDir)
			os.Exit(1)
		default:
			log.Error("unable to read config", "error", err)
			os.Exit(1)
		}
	}
	var c Config
	cfg.Unmarshal(&c)
	return c
}
