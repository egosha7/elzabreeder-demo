package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"net"
	"regexp"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// Config - структура конфигурации приложения
type Config struct {
	Addr          string `env:"SERVER_ADDRESS" json:"server_address"`                 // Адрес сервера
	BaseURL       string `env:"BASE_URL" json:"base_url"`                             // Базовый адрес результирующего сокращенного URL
	DataBase      string `env:"DATABASE_DSN" json:"database_dsn"`                     // Адрес базы данных PostgresSQL
	UriMongoDB    string `env:"DATABASE_MONGO" json:"database_mongoDBUri"`            // Адрес базы данных MongoDB
	MongoDBName   string `env:"DATABASE_MONGODBNAME" json:"database_mongoDBName"`     // Название базы данных MongoDB
	RedisAddress  string `env:"DATABASE_REDIS" json:"database_redisAddress"`          // Адрес базы данных Redis
	RedisPassword string `env:"DATABASE_REDISPASSWORD" json:"database_redisPassword"` // Пароль базы данных Redis
	RedisDBName   int    `env:"DATABASE_REDISDBNAME" json:"database_RedisDBName"`     // Имя базы данных Redis
}

// Default - функция для создания новой конфигурации со значениями по умолчанию
func Default() *Config {
	return &Config{
		Addr:          "192.168.3.69:8080",
		BaseURL:       "http://192.168.3.69:8080",
		DataBase:      "postgres://postgres:egosha@localhost:5432/ElzaBreeder",
		UriMongoDB:    "mongodb://localhost:27017",
		RedisAddress:  "localhost:6379",
		RedisPassword: "",
		RedisDBName:   1,
	}
}

// OnFlag - функция для чтения значений из флагов командной строки и записи их в структуру Config
func OnFlag(logger *zap.Logger) *Config {
	defaultValue := Default()

	// Инициализация флагов командной строки
	config := Config{}
	flag.StringVar(&config.Addr, "a", defaultValue.Addr, "HTTP-адрес сервера")
	flag.StringVar(&config.BaseURL, "b", defaultValue.BaseURL, "Базовый адрес результирующего сокращенного URL")
	flag.StringVar(&config.DataBase, "d", defaultValue.DataBase, "Адрес базы данных PostgresSQL")
	flag.StringVar(&config.UriMongoDB, "c", defaultValue.UriMongoDB, "Адрес базы данных MongoDB")
	flag.StringVar(&config.RedisAddress, "r", defaultValue.RedisAddress, "Адрес базы данных Redis")
	flag.StringVar(&config.RedisPassword, "rp", defaultValue.RedisPassword, "Пароль базы данных Redis")
	flag.IntVar(&config.RedisDBName, "rn", defaultValue.RedisDBName, "Имя базы данных Redis")
	flag.Parse()

	godotenv.Load()

	// Парсинг переменных окружения в структуру Config
	if err := env.Parse(&config); err != nil {
		logger.Error("Ошибка при парсинге переменных окружения", zap.Error(err))
	}

	// Проверка корректности введенных значений флагов
	if _, _, err := net.SplitHostPort(config.Addr); err != nil {
		panic(err)
	}
	if matched, _ := regexp.MatchString(`^https?://[^\s/$.?#].[^\s]*$`, config.BaseURL); !matched {
		panic("Invalid base URL")
	}

	return &config
}
