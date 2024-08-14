package initial

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/egosha7/site-go/internal/config"
	"github.com/egosha7/site-go/internal/db"
	"github.com/egosha7/site-go/internal/handlers"
	mailer2 "github.com/egosha7/site-go/internal/mailer"
	"github.com/egosha7/site-go/internal/repository"
	"github.com/egosha7/site-go/internal/service"
	"go.uber.org/zap"
)

const (
	smtpHost      = "smtp.yandex.ru"
	smtpPort      = 587
	smtpUser      = "egor.sharik"
	smtpPass      = "ldqqtslabxoqwppm"
	smtpFromEmail = "egor.sharik@yandex.ru"
)

func Initial(logger *zap.Logger) (*config.Config, *handlers.Handler) {
	logger.Info("Start initial...")
	// Проверка конфигурации из флагов и переменных окружения.
	cfg := config.OnFlag(logger)

	// Создание пула подключений
	pool, err := db.ConnectToPostgresDB(cfg)
	if err != nil {
		logger.Fatal("Failed connect (Postgres)", zap.Error(err))
	} else {
		logger.Info("Connected to Postgres")
	}

	// Подключение к базе данных Mongo.
	clientMongo, err := db.ConnectToMongoDB(cfg.UriMongoDB)
	if err != nil {
		logger.Fatal("Failed connect (Mongo)", zap.Error(err))
	} else {
		logger.Info("Connected to MongoDB")
	}

	clientRedis, err := db.InitRedisClient(cfg.RedisAddress, cfg.RedisPassword, cfg.RedisDBName)
	if err != nil {
		logger.Fatal("Failed connect (Redis)", zap.Error(err))
	} else {
		logger.Info("Connected to Redis")
	}
	clientRedis.FlushAll(context.Background()).Err()

	s3Client, err := db.InitS3Client()
	if err != nil {
		logger.Fatal("Failed to initialize S3 client", zap.Error(err))
	} else {
		logger.Info("Connected to S3")
	}

	// Проверка подключения: Получение списка бакетов
	result, err := s3Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		logger.Fatal("Failed to list buckets: ", zap.Error(err))
	}

	logger.Info("Successfully connected to S3. Buckets:")
	for _, bucket := range result.Buckets {
		logger.Info("Successfully connected to S3. Buckets:")
		fmt.Printf("* %s\n", aws.ToString(bucket.Name))
	}

	// Проверка подключения: Получение списка объектов в бакете
	objectResult, err := s3Client.ListObjectsV2(
		context.TODO(), &s3.ListObjectsV2Input{
			Bucket: aws.String(db.BucketName),
		},
	)
	if err != nil {
		logger.Fatal("Failed to list objects in bucket:", zap.Error(err))
	}

	logger.Info("Objects in bucket " + db.BucketName + " :")
	for _, object := range objectResult.Contents {
		fmt.Printf("* %s\n", aws.ToString(object.Key))
	}

	// Создание хранилища
	repo := repository.NewRepository(pool, clientMongo, clientRedis, logger, s3Client, db.BucketName)
	services := service.NewUserService(repo, logger)
	h := handlers.NewHandler(services, logger)
	mailer := mailer2.NewMailer(
		smtpHost, smtpPort, smtpUser, smtpPass, smtpFromEmail,
	)

	// Проверяем подключение к SMTP-серверу
	err = mailer.CheckSMTPConnection()
	if err != nil {
		logger.Fatal("Failed to connect to SMTP server: ", zap.Error(err))
	} else {
		logger.Info("Successfully connected to SMTP server")
	}
	logger.Info("Successfully initialization!")
	return cfg, h
}
