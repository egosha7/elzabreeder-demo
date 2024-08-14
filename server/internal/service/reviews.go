package service

import (
	"fmt"
	"github.com/egosha7/site-go/internal/domain"
	"go.uber.org/zap"
	"mime/multipart"
)

// ReviewsGet получает отзывы.
func (s *ServiceImpl) ReviewsGet(idReview string, checked bool) ([]domain.Feedback, map[int]string, error) {
	cacheKeyReviews := fmt.Sprintf("reviews:%s:%t", idReview, checked)
	cacheKeyPuppyNames := "puppyNames"

	cachedReviews, err := s.Repository.RedisRepository.GetReviews(cacheKeyReviews)
	if err == nil && cachedReviews != nil {
		cachedPuppyNames, err := s.Repository.RedisRepository.GetPuppyNames(cacheKeyPuppyNames)
		if err == nil && cachedPuppyNames != nil {
			return cachedReviews, cachedPuppyNames, nil
		}
	}

	reviews, err := s.Repository.PostgresRepository.ReviewsGet(idReview, checked)
	if err != nil {
		s.Logger.Error("Ошибка сервиса", zap.Error(err))
		return nil, nil, err
	}

	puppyNames, err := s.Repository.PostgresRepository.ReviewsPuppyNameGet()
	if err != nil {
		return nil, nil, err
	}

	go func() {
		err := s.Repository.RedisRepository.SetReviews(cacheKeyReviews, reviews)
		if err != nil {
			s.Logger.Error("Ошибка кеширования отзывов", zap.Error(err))
		}

		err = s.Repository.RedisRepository.SetPuppyNames(cacheKeyPuppyNames, puppyNames)
		if err != nil {
			s.Logger.Error("Ошибка кеширования имен щенков", zap.Error(err))
		}
	}()

	return reviews, puppyNames, nil
}

// FeedbackChangeChecked меняет состояние архива отзыва.
func (s *ServiceImpl) FeedbackChangeChecked(feedbackID string, checked string) error {
	s.Repository.RedisRepository.FlushAll()
	return s.Repository.FeedbackChangeChecked(feedbackID, checked)
}

// FeedbackDelete удаляет информацию об отзыве.
func (s *ServiceImpl) FeedbackDelete(feedbackID string) error {
	s.Repository.RedisRepository.FlushAll()
	imgUrls, err := s.Repository.FeedbackDelete(feedbackID)
	for _, imgUrl := range imgUrls {
		err = s.Repository.S3Repository.DeleteFromS3(imgUrl)
		if err != nil {
			return fmt.Errorf("unable to delete from image URLs: %w", err)
		}
	}
	return err
}

// FeedbackUpdate обновляет информацию об отзыве.
func (s *ServiceImpl) FeedbackUpdate(feedback *domain.Feedback, fileHeaders []*multipart.FileHeader) error {
	s.Repository.RedisRepository.FlushAll()
	newUrls, err := s.Repository.S3Repository.PutInS3(feedback.Name, fileHeaders, 1.0/1.0, 1000, 1000)
	if err != nil {
		return err
	}
	// Объединение старых и новых URL-адресов
	feedback.Urls = append(feedback.Urls, newUrls...)
	currentUrlSet, newUrlSet, err := s.Repository.FeedbackUpdate(feedback)
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

// FeedbackGet получает отзыв о щенке.
func (s *ServiceImpl) FeedbackGet(idPuppy, verified string) (*domain.Feedback, error) {
	cacheKey := fmt.Sprintf("feedback:%s:verified:%s", idPuppy, verified)

	cachedFeedback, err := s.Repository.RedisRepository.GetFeedback(cacheKey)
	if err == nil && cachedFeedback != nil {
		return cachedFeedback, nil
	}

	feedback, err := s.Repository.PostgresRepository.FeedbackGet(idPuppy, verified)
	if err != nil {
		return nil, err
	}

	go func() {
		err := s.Repository.RedisRepository.SetFeedback(cacheKey, feedback)
		if err != nil {
			s.Logger.Error("Ошибка кеширования отзыва", zap.Error(err))
		}
	}()

	return feedback, nil
}

// FeedbackAdd добавляет отзыв.
func (s *ServiceImpl) FeedbackAdd(feedback *domain.Feedback, fileHeaders []*multipart.FileHeader) error {
	s.Repository.RedisRepository.FlushAll()
	urls, err := s.Repository.PutInS3(feedback.Name, fileHeaders, 1.0/1.0, 1000, 1000)
	if err != nil {
		return err
	}
	feedback.Urls = urls
	println(feedback.Urls)
	return s.Repository.FeedbackAdd(feedback)
}
