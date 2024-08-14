package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"log"
)

// MongosRepository представляет интерфейс для работы с MongoDB.
type MongosRepository interface {
	AddEmail(email string) error
	GetSubscribers() ([]string, error)
}

// MongosRepo представляет репозиторий для работы с MongoDB.
type MongosRepo struct {
	logger      *zap.Logger
	mongoClient *mongo.Client
}

// AddEmail проверяет валидность пользователя.
func (r *MongosRepo) AddEmail(email string) error {
	// Получаем коллекцию "emails" из базы данных
	collection := r.mongoClient.Database("ElzaBreeder").Collection("emails")

	// Создаем документ для вставки в коллекцию
	document := bson.D{
		{Key: "email", Value: email},
	}

	// Вставляем документ в коллекцию
	_, err := collection.InsertOne(context.Background(), document)
	if err != nil {
		log.Printf("Failed to insert email into MongoDB: %v", err)
		return err
	}

	r.logger.Info("Email successfully inserted into MongoDB" + ": " + email)

	return nil
}

func (r *MongosRepo) GetSubscribers() ([]string, error) {
	collection := r.mongoClient.Database("ElzaBreeder").Collection("emails")

	var subscribers []string
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var result struct {
			Email string `bson:"email"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		subscribers = append(subscribers, result.Email)
	}

	return subscribers, nil
}
