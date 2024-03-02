package main

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var GlobalConfig *Config

type environment string

const (
	// EnvLocal represents the local environment
	EnvLocal environment = "local"

	// EnvTest represents the test environment
	EnvTest environment = "test"

	// EnvDevelop represents the development environment
	EnvDevelop environment = "dev"

	// EnvProduction represents the production environment
	EnvProduction environment = "prod"
)

type (
	// Config stores complete configuration
	Config struct {
		App      AppConfig        `mapstructure:"app"`
		Database DatabaseConfig   `mapstructure:"db"`
		Mail     MailConfig       `mapstructure:"mail-ingest"`
		Encrypt  EncryptionConfig `mapstructure:"auth-cli"`
	}

	AppConfig struct {
		Name        string      `mapstructure:"name"`
		Runtime     environment `mapstructure:"runtime"`
		Version     string      `mapstructure:"version"`
		ScheduleSec int         `mapstructure:"schedule_sec"`
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
		Accounts *[]MailAccountConfig `mapstructure:"accounts"`
	}

	MailAccountConfig struct {
		Hostname    string              `mapstructure:"host"`
		Port        int                 `mapstructure:"port"`
		EncUser     string              `mapstructure:"user"`
		EncPassword string              `mapstructure:"password"`
		Prune       MailboxActionConfig `mapstructure:"prune"`
		BackupOnly  MailboxActionConfig `mapstructure:"backup"`
	}
	MailboxActionConfig struct {
		ThresholdDays int      `mapstructure:"threshold_days,omitempty"`
		Folders       []string `mapstructure:"folders"`
	}
)

// NewConfig returns app config.
func InitConfig() *Config {

	cfg := viper.New()
	cfg.AddConfigPath("./conf/")
	cfg.AddConfigPath("/home/zee/sharp/newsSummarizer/archive/conf")
	cfg.AddConfigPath("/etc/archive")
	cfg.SetConfigType("yaml")
	cfg.SetConfigName(".settings")

	err := cfg.ReadInConfig()

	contextDir, _ := os.Getwd()
	log.Info().Msgf("Looking for config from context %s", contextDir)
	if err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			log.Fatal().Err(err).Str("working_dir", contextDir).
				Msg("No Config file found")
		default:
			log.Fatal().Err(err).Msg("Unable to read config.")
		}
	}

	cfg.Unmarshal(&GlobalConfig)

	log.Info().Msg("Successfully initalized global configuration.")
	return GlobalConfig
}

func (cfg *Config) IsInDev() bool {
	return cfg.App.Runtime == EnvDevelop
}
