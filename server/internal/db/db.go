package db

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/egosha7/site-go/internal/config"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ConnectToPostgresDB устанавливает соединение с базой данных на основе конфигурации.
// Возвращает соединение (pgx. Conn) и ошибку, если возникает ошибка при подключении.
func ConnectToPostgresDB(cfg *config.Config) (*pgxpool.Pool, error) {
	// Парсинг конфигурации для пула подключений
	configDB, err := pgxpool.ParseConfig(cfg.Postgres.DSN)
	if err != nil {
		return nil, err
	}

	// Установка максимального количества соединений в пуле
	configDB.MaxConns = 1000

	// Создание пула подключений
	pool, err := pgxpool.ConnectConfig(context.Background(), configDB)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

// PingDB выполняет пинг базы данных и отправляет статус в HTTP-ответ.
func PingDB(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {
	err := conn.Ping(context.Background())
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ConnectToMongoDB подключается к базе данных MongoDB и возвращает клиента
func ConnectToMongoDB(uri string) (*mongo.Client, error) {
	// Устанавливаем параметры подключения к MongoDB
	clientOptions := options.Client().ApplyURI(uri)

	// Подключаемся к серверу MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Printf("Failed to connect to MongoDB: %v", err)
		return nil, err
	}

	// Проверяем, что соединение действительно установлено
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Printf("Failed to ping MongoDB: %v", err)
		return nil, err
	}

	return client, nil
}

func InitS3Client(cfgS3 config.S3Config) (*s3.Client, error) {
	awsCfg := aws.Config{
		Region:      cfgS3.Region,
		Credentials: credentials.NewStaticCredentialsProvider(cfgS3.AccessKey, cfgS3.SecretKey, ""),
		EndpointResolverWithOptions: aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				if service == s3.ServiceID {
					return aws.Endpoint{
						URL:           cfgS3.URL,
						SigningRegion: cfgS3.Region,
					}, nil
				}
				return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
			},
		),
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return client, nil
}

var ctx = context.Background()

// InitRedisClient инициализация Redis клиента
func InitRedisClient(addr, password string, db int) (*redis.Client, error) {
	rdb := redis.NewClient(
		&redis.Options{
			Addr:     addr,     //
			Password: password, //
			DB:       db,       //
		},
	)

	// Тест подключения
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	fmt.Println("Connected to Redis:", pong)

	return rdb, nil
}
