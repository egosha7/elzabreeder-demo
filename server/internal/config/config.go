package config

import (
	"fmt"

	"github.com/caarlos0/env/v6"
	"gopkg.in/yaml.v3"

	"os"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// Config - структура конфигурации приложения
type Config struct {
	Host     HostConfig     `yaml:"host"`
	Postgres PostgresConfig `yaml:"postgres"`
	Mongo    MongoConfig    `yaml:"mongo"`
	Redis    RedisConfig    `yaml:"redis"`
	S3       S3Config       `yaml:"s3"`
}

type HostConfig struct {
	Addr     string `yaml:"addr" env:"SERVER_ADDRESS" json:"addr"`      // Адрес сервера
	BaseURL  string `yaml:"base_url" env:"BASE_URL" json:"base_url"`    // Базовый адрес результирующего сокращенного URL
	CertFile string `yaml:"cert_file" env:"CERT_FILE" json:"cert_file"` // Путь к файлу сертификата
	KeyFile  string `yaml:"key_file" env:"KEY_FILE" json:"key_file"`    // Путь к файлу ключа
}

type PostgresConfig struct {
	DSN string `yaml:"dsn" env:"DATABASE_DSN" json:"dsn"`
}

type MongoConfig struct {
	DSN string `yaml:"dsn" env:"MONGO_DSN" json:"dsn"`
}

type RedisConfig struct {
	Address  string `yaml:"address" env:"DATABASE_REDIS" json:"address"`           // Адрес базы данных Redis
	Password string `yaml:"password" env:"DATABASE_REDISPASSWORD" json:"password"` // Пароль базы данных Redis
	DBName   int    `yaml:"db_name" env:"DATABASE_REDISDBNAME" json:"db_name"`     // Имя базы данных Redis
}

type S3Config struct {
	URL       string `yaml:"url" env:"S3_URL" json:"url"`
	Region    string `yaml:"region" env:"S3_REGION" json:"region"`
	AccessKey string `yaml:"access_key" env:"S3_ACCESS_KEY" json:"access_key"`
	SecretKey string `yaml:"secret_key" env:"S3_SECRET_KEY" json:"secret_key"`
	Bucket    string `yaml:"bucket" env:"S3_BUCKET" json:"bucket"`
}

// Default - функция для создания новой конфигурации со значениями по умолчанию
func Default() *Config {
	return &Config{
		Host: HostConfig{
			Addr:     "192.168.3.69:8080",
			BaseURL:  "http://192.168.3.69:8080",
			CertFile: "",
			KeyFile:  "",
		},
		Postgres: PostgresConfig{
			DSN: "postgres://postgres:egosha@localhost:5432/ElzaBreeder",
		},
		Mongo: MongoConfig{
			DSN: "mongodb://localhost:27017",
		},
		Redis: RedisConfig{
			Address:  "localhost:6379",
			Password: "",
			DBName:   0,
		},
		S3: S3Config{
			URL:       "http://192.168.3.69:9000",
			Region:    "us-east-1",
			AccessKey: "minioadmin",
			SecretKey: "minioadmin",
			Bucket:    "elzabreeder",
		},
	}
}

func Load(path string, logger *zap.Logger) (*Config, error) {
	cfg := Default()

	if path != "" {
		file, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		if err := yaml.Unmarshal(file, cfg); err != nil {
			return nil, err
		}

		fmt.Println(cfg)
	}

	godotenv.Load()

	if err := env.Parse(cfg); err != nil {
		logger.Error("env parse error", zap.Error(err))
	}

	return cfg, nil
}
