package main

import (
	"flag"
	"net/url"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config config data for app.
type Config struct {
	// Адрес запуска HTTP-сервера. Пример: localhost:8080
	ServerAddress string `env:"SERVER_ADDRESS" mapstructure:"server_address"` // address to start the server
	GRPCAddress   string `env:"GRPC_ADDRESS" mapstructure:"grpc_address"`
	// Базовый адрес результирующего сокращённого URL
	// Требования:
	//     - Должен быть указан протокол (по умолчанию автоматически добавится http://): http/https
	//     - Путь URL Path должен быть (по-умолчанию автоматически добавится /)
	// Пример: http://localhost:8080/blablabla
	BaseURL         string `env:"BASE_URL" mapstructure:"base_url"` // short URLs will be returned with this host
	LogLevel        string `env:"LOG_LEVEL" mapstructure:"log_level"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" mapstructure:"file_storage_path"`
	DatabaseDSN     string `env:"DATABASE_DSN" mapstructure:"database_dsn"`
	EnableHTTPS     bool   `env:"ENABLE_HTTPS" mapstructure:"enable_https"`
	Config          string `env:"CONFIG"`
	TrustedSubnet   string `env:"TRUSTED_SUBNET" mapstructure:"trusted_subnet"`
}

func readConfigFile(c *Config) error {
	v := viper.New()

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	err := v.BindPFlag("server_address", pflag.Lookup("a"))
	if err != nil {
		return err
	}
	err = v.BindPFlag("grpc_address", pflag.Lookup("g"))
	if err != nil {
		return err
	}
	err = v.BindPFlag("base_url", pflag.Lookup("b"))
	if err != nil {
		return err
	}
	err = v.BindPFlag("log_level", pflag.Lookup("l"))
	if err != nil {
		return err
	}
	err = v.BindPFlag("file_storage_path", pflag.Lookup("f"))
	if err != nil {
		return err
	}
	err = v.BindPFlag("database_dsn", pflag.Lookup("d"))
	if err != nil {
		return err
	}
	err = v.BindPFlag("enable_https", pflag.Lookup("s"))
	if err != nil {
		return err
	}
	err = v.BindPFlag("trusted_subnet", pflag.Lookup("t"))
	if err != nil {
		return err
	}

	v.SetConfigFile(c.Config)
	v.AutomaticEnv()

	err = v.ReadInConfig()
	if err != nil {
		return err
	}

	err = v.Unmarshal(c)
	if err != nil {
		return err
	}

	return nil
}

// NewConfig create config
func NewConfig() (*Config, error) {
	c := &Config{}

	flag.StringVar(&c.ServerAddress, "a", "", "Server address")
	flag.StringVar(&c.GRPCAddress, "g", "", "GRPC address")
	flag.StringVar(&c.BaseURL, "b", "", "Base URL")
	flag.StringVar(&c.LogLevel, "l", "", "Log level")
	flag.StringVar(&c.FileStoragePath, "f", "", "File storage path")
	flag.StringVar(&c.DatabaseDSN, "d", "", "Database DSN")
	flag.BoolVar(&c.EnableHTTPS, "s", false, "Enable HTTPS")
	flag.StringVar(&c.Config, "c", "", "Config path")
	flag.StringVar(&c.TrustedSubnet, "t", "", "Trusted subnet")
	flag.Parse()

	foundFlagFileStoragePath := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "f" {
			foundFlagFileStoragePath = true
		}
	})

	err := env.Parse(c)
	if err != nil {
		return nil, err
	}

	if c.Config != "" {
		err = readConfigFile(c)
		if err != nil {
			return nil, err
		}
	}

	_, foundEnvFileStoragePath := os.LookupEnv("FILE_STORAGE_PATH")
	_, foundEnvEnableHTTPS := os.LookupEnv("ENABLE_HTTPS")

	if foundEnvEnableHTTPS {
		c.EnableHTTPS = true
	}

	// Если не ввели -a, -b, -l, -f то значения по-умолчанию
	if c.ServerAddress == "" {
		c.ServerAddress = Addr
	}
	if c.GRPCAddress == "" {
		c.GRPCAddress = GRPCAddr
	}
	if c.BaseURL == "" {
		c.BaseURL = ResultAddrPrefix
	}
	if c.LogLevel == "" {
		c.LogLevel = LogLevel
	}
	if !foundFlagFileStoragePath && !foundEnvFileStoragePath {
		c.FileStoragePath = URLsFileStoragePath
	}

	_, err = url.ParseRequestURI(c.BaseURL)
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(c.BaseURL, "http://") && !c.EnableHTTPS {
		c.BaseURL = "http://" + c.BaseURL
	} else if !strings.HasPrefix(c.BaseURL, "https://") && c.EnableHTTPS {
		c.BaseURL = "https://" + c.BaseURL
	}
	if !strings.HasSuffix(c.BaseURL, "/") {
		c.BaseURL += "/"
	}

	return c, nil
}
