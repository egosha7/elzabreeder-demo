package service

import (
	"fmt"
	"github.com/egosha7/site-go/internal/domain"
	"mime/multipart"
)

// DogsGet получает список собак.
func (s *ServiceImpl) DogsGet(chocolates, genders []string, id string, archived bool) ([]domain.Dog, error) {
	// Здесь может быть ваша бизнес-логика
	return s.Repository.DogsGet(chocolates, genders, id, archived)
}

// DogChangeArchived меняет состояние архива собаки.
func (s *ServiceImpl) DogChangeArchived(dogID string, archived string) error {
	s.Repository.RedisRepository.FlushAll()
	return s.Repository.DogChangeArchived(dogID, archived)
}

// DogAdd добавляет информацию о собаке.
func (s *ServiceImpl) DogAdd(dog *domain.Dog, fileHeaders []*multipart.FileHeader) error {
	s.Repository.RedisRepository.FlushAll()
	urls, err := s.Repository.PutInS3(dog.Name, fileHeaders, 3.0/2.0, 1200, 800)
	if err != nil {
		return err
	}
	dog.Urls = urls
	println(dog.Urls)
	return s.Repository.DogAdd(dog)
}

// DogUpdate обновляет информацию о собаке.
func (s *ServiceImpl) DogUpdate(dog *domain.Dog, fileHeaders []*multipart.FileHeader) error {
	s.Repository.RedisRepository.FlushAll()
	newUrls, err := s.Repository.PutInS3(dog.Name, fileHeaders, 3.0/2.0, 1200, 800)
	if err != nil {
		return err
	}
	// Объединение старых и новых URL-адресов
	dog.Urls = append(dog.Urls, newUrls...)
	currentUrlSet, newUrlSet, err := s.Repository.DogUpdate(dog)
	if err != nil {
		return err
	}
	// Удаление старых файлов, которые больше не существуют
	for url := range currentUrlSet {
		if _, exists := newUrlSet[url]; !exists {
			// Удаление файла из S3
			err = s.Repository.S3Repository.DeleteFromS3(url)
			if err != nil {
				return fmt.Errorf("failed to delete image from S3: %w", err)
			}
		}
	}
	return err
}
