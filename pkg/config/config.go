package config

import (
	"fmt"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	Server   *ServerConfig   `mapstructure:"server" validate:"required"`
	Logging  *LoggingConfig  `mapstructure:"logging"`
	Database *DatabaseConfig `mapstructure:"database" validate:"required"`
}

type ServerConfig struct {
	Port int `mapstructure:"port" validate:"required,min=1,max=65535"`
}

type LoggingConfig struct {
	Level string `mapstructure:"level" validate:"oneof=debug info warn error"`

	Rotation *struct {
		Name               string `mapstructure:"name" validate:"required"`
		MaxSize            int    `mapstructure:"max_size" validate:"min=0"`
		MaxBackups         int    `mapstructure:"max_backups" validate:"min=0"`
		MaxAge             int    `mapstructure:"max_age" validate:"min=0"`
		DuplicateToConsole bool   `mapstructure:"duplicate_to_console" validate:"omitempty"`
	} `mapstructure:"rotation"`
}

type DatabaseConfig struct {
	URL            string `mapstructure:"url" validate:"required"`
	MaxConnections int    `mapstructure:"max_connections" validate:"required,min=1,max=100"`
}

var (
	cfg  *Config
	once sync.Once
)

func init() {
	var loadErr error
	once.Do(func() {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./configs")

		viper.SetEnvPrefix("APP")
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		viper.AutomaticEnv()

		_ = viper.BindEnv("server.port")
		_ = viper.BindEnv("database.driver")
		_ = viper.BindEnv("database.url")

		if err := viper.ReadInConfig(); err != nil {
			loadErr = fmt.Errorf("failed to read config: %w", err)
			return
		}

		cfg = &Config{}
		if err := viper.Unmarshal(&cfg); err != nil {
			loadErr = fmt.Errorf("failed to unmarshal config: %w", err)
			return
		}

		validate := validator.New(validator.WithRequiredStructEnabled())
		if err := validate.Struct(cfg); err != nil {
			loadErr = fmt.Errorf("config validation failed: %w", err)
			return
		}
	})
	if loadErr != nil {
		panic(fmt.Sprintf("Config init failed: %v", loadErr))
	}
}

func Get() *Config {
	if cfg == nil {
		panic("config not initialized")
	}
	return cfg
}
