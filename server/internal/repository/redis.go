package repository

import (
	"context"
	"encoding/json"
	"github.com/egosha7/site-go/internal/domain"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"time"
)

// RedisRepository представляет интерфейс для работы с Redis.
type RedisRepository interface {
	GetPuppies(cacheKey string) (*CachedPuppies, error)
	SetPuppies(cacheKey string, cachedPuppies *CachedPuppies) error
	GetPuppyReviews(cacheKey string) (map[int]int, error)
	SetPuppyReviews(cacheKey string, reviews map[int]int) error
	GetReviews(cacheKey string) ([]domain.Feedback, error)
	SetReviews(cacheKey string, reviews []domain.Feedback) error
	GetFeedback(cacheKey string) (*domain.Feedback, error)
	SetFeedback(cacheKey string, feedback *domain.Feedback) error
	GetPuppyNames(cacheKey string) (map[int]string, error)
	SetPuppyNames(cacheKey string, puppyNames map[int]string) error
	GetPuppy(cacheKey string) (*domain.Puppy, error)
	SetPuppy(cacheKey string, puppy *domain.Puppy) error
	GetDog(cacheKey string) (*domain.Dog, error)
	SetDog(cacheKey string, dog *domain.Dog) error
	FlushAll()
}

// RedisRepo представляет репозиторий для работы с Redis.
type RedisRepo struct {
	client *redis.Client
	logger *zap.Logger
}

type CachedPuppies struct {
	Puppies    []domain.Puppy
	TotalPages int
}

func (r *RedisRepo) GetPuppies(cacheKey string) (*CachedPuppies, error) {
	r.logger.Info("Start get cache GetPuppies")
	val, err := r.client.Get(context.Background(), cacheKey).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var cachedPuppies CachedPuppies
	err = json.Unmarshal([]byte(val), &cachedPuppies)
	if err != nil {
		return nil, err
	}
	return &cachedPuppies, nil
}

func (r *RedisRepo) SetPuppies(cacheKey string, cachedPuppies *CachedPuppies) error {
	r.logger.Info("Start set cache SetPuppies")
	data, err := json.Marshal(cachedPuppies)
	if err != nil {
		return err
	}

	err = r.client.Set(context.Background(), cacheKey, data, time.Hour).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisRepo) GetPuppyReviews(cacheKey string) (map[int]int, error) {
	r.logger.Info("Start get cache GetPuppyReviews")
	val, err := r.client.Get(context.Background(), cacheKey).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var cachedReviews map[int]int
	err = json.Unmarshal([]byte(val), &cachedReviews)
	if err != nil {
		return nil, err
	}
	return cachedReviews, nil
}

func (r *RedisRepo) SetPuppyReviews(cacheKey string, reviews map[int]int) error {
	r.logger.Info("Start set cache SetPuppyReviews")
	data, err := json.Marshal(reviews)
	if err != nil {
		return err
	}

	err = r.client.Set(context.Background(), cacheKey, data, time.Hour).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisRepo) GetReviews(cacheKey string) ([]domain.Feedback, error) {
	r.logger.Info("Start get cache GetReviews")
	val, err := r.client.Get(context.Background(), cacheKey).Result()
	if err == redis.Nil {
		return nil, nil // Данных нет в кеше
	} else if err != nil {
		return nil, err
	}

	var reviews []domain.Feedback
	err = json.Unmarshal([]byte(val), &reviews)
	if err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *RedisRepo) SetReviews(cacheKey string, reviews []domain.Feedback) error {
	r.logger.Info("Start set cache SetReviews")
	data, err := json.Marshal(reviews)
	if err != nil {
		return err
	}

	err = r.client.Set(context.Background(), cacheKey, data, time.Hour).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisRepo) GetFeedback(cacheKey string) (*domain.Feedback, error) {
	r.logger.Info("Start get cache GetFeedback")
	val, err := r.client.Get(context.Background(), cacheKey).Result()
	if err == redis.Nil {
		return nil, nil // Данных нет в кеше
	} else if err != nil {
		return nil, err
	}

	var feedback domain.Feedback
	err = json.Unmarshal([]byte(val), &feedback)
	if err != nil {
		return nil, err
	}
	return &feedback, nil
}

func (r *RedisRepo) SetFeedback(cacheKey string, feedback *domain.Feedback) error {
	r.logger.Info("Start set cache SetFeedback")
	data, err := json.Marshal(feedback)
	if err != nil {
		return err
	}

	err = r.client.Set(context.Background(), cacheKey, data, time.Hour).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisRepo) GetPuppyNames(cacheKey string) (map[int]string, error) {
	r.logger.Info("Start get cache GetPuppyNames")
	val, err := r.client.Get(context.Background(), cacheKey).Result()
	if err == redis.Nil {
		return nil, nil // Данных нет в кеше
	} else if err != nil {
		return nil, err
	}

	var puppyNames map[int]string
	err = json.Unmarshal([]byte(val), &puppyNames)
	if err != nil {
		return nil, err
	}
	return puppyNames, nil
}

func (r *RedisRepo) SetPuppyNames(cacheKey string, puppyNames map[int]string) error {
	r.logger.Info("Start set cache SetPuppyNames")
	data, err := json.Marshal(puppyNames)
	if err != nil {
		return err
	}

	err = r.client.Set(context.Background(), cacheKey, data, time.Hour).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisRepo) GetPuppy(cacheKey string) (*domain.Puppy, error) {
	r.logger.Info("Start get cache GetPuppy")
	val, err := r.client.Get(context.Background(), cacheKey).Result()
	if err == redis.Nil {
		return nil, nil // Данных нет в кеше
	} else if err != nil {
		return nil, err
	}

	var puppy domain.Puppy
	err = json.Unmarshal([]byte(val), &puppy)
	if err != nil {
		return nil, err
	}
	return &puppy, nil
}

func (r *RedisRepo) SetPuppy(cacheKey string, puppy *domain.Puppy) error {
	r.logger.Info("Start set cache SetPuppy")
	data, err := json.Marshal(puppy)
	if err != nil {
		return err
	}

	err = r.client.Set(context.Background(), cacheKey, data, time.Hour).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisRepo) GetDog(cacheKey string) (*domain.Dog, error) {
	r.logger.Info("Start get cache GetDog")
	val, err := r.client.Get(context.Background(), cacheKey).Result()
	if err == redis.Nil {
		return nil, nil // Данных нет в кеше
	} else if err != nil {
		return nil, err
	}

	var dog domain.Dog
	err = json.Unmarshal([]byte(val), &dog)
	if err != nil {
		return nil, err
	}
	return &dog, nil
}

func (r *RedisRepo) SetDog(cacheKey string, dog *domain.Dog) error {
	r.logger.Info("Start set cache SetDog")
	data, err := json.Marshal(dog)
	if err != nil {
		return err
	}

	err = r.client.Set(context.Background(), cacheKey, data, time.Hour).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisRepo) FlushAll() {
	r.logger.Info("Start Flush All")
	r.client.FlushAll(context.Background())
	return
}
