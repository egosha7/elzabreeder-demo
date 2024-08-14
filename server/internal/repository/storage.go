package repository

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type Repository struct {
	PostgresRepository
	RedisRepository
	S3Repository
	MongosRepository
}

// NewRepository создает новый экземпляр Repository.
func NewRepository(pool *pgxpool.Pool, mongoClient *mongo.Client, redisClient *redis.Client, logger *zap.Logger, s3Client *s3.Client, bucket string) *Repository {
	return &Repository{
		PostgresRepository: &PostgresRepo{
			pool:   pool,
			logger: logger,
		},
		RedisRepository: &RedisRepo{
			client: redisClient,
			logger: logger,
		},
		S3Repository: &S3Repo{
			logger:   logger,
			S3Client: s3Client,
			bucket:   bucket,
		},
		MongosRepository: &MongosRepo{
			logger:      logger,
			mongoClient: mongoClient,
		},
	}
}
