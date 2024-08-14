package service

import (
	"fmt"
	"github.com/egosha7/site-go/internal/domain"
	"github.com/egosha7/site-go/internal/repository"
	"go.uber.org/zap"
	"log"
	"mime/multipart"
	"strconv"
)

// PuppiesGet получает информацию о щенках.
func (s *ServiceImpl) PuppiesGet(chocolates, genders []string, idPuppy, readyToMove string, page int, archived bool) ([]domain.Puppy, map[int]int, int, error) {
	cacheKeyPuppies := fmt.Sprintf(
		"puppies:chocolates:%v:genders:%v:idPuppy:%s:readyToMove:%s:archived:%t:page:%d", chocolates, genders, idPuppy,
		readyToMove, archived, page,
	)
	cacheKeyReviews := "puppyReviews"

	cachedPuppies, err := s.Repository.RedisRepository.GetPuppies(cacheKeyPuppies)
	if err == nil && cachedPuppies != nil {
		cachedReviews, err := s.Repository.RedisRepository.GetPuppyReviews(cacheKeyReviews)
		if err == nil && cachedReviews != nil {
			return cachedPuppies.Puppies, cachedReviews, cachedPuppies.TotalPages, nil
		}
	}

	puppies, err := s.Repository.PostgresRepository.PuppiesGet(chocolates, genders, idPuppy, readyToMove, archived)
	if err != nil {
		s.Logger.Error("Ошибка получения данных щенков", zap.Error(err))
		return nil, nil, 0, err
	}

	reviews, err := s.Repository.PostgresRepository.PuppiesWithReviewsGet()
	if err != nil {
		s.Logger.Error("Ошибка получения данных отзывов", zap.Error(err))
		return nil, nil, 0, err
	}

	pagedPuppies, totalPages, err := s.GetPagedPuppies(puppies, page, 2)
	if err != nil {
		return nil, nil, 0, err
	}

	go func() {
		err := s.Repository.RedisRepository.SetPuppies(
			cacheKeyPuppies, &repository.CachedPuppies{
				Puppies:    pagedPuppies,
				TotalPages: totalPages,
			},
		)
		if err != nil {
			s.Logger.Error("Ошибка кеширования щенков", zap.Error(err))
		}

		err = s.Repository.RedisRepository.SetPuppyReviews(cacheKeyReviews, reviews)
		if err != nil {
			s.Logger.Error("Ошибка кеширования отзывов", zap.Error(err))
		}
	}()

	return pagedPuppies, reviews, totalPages, nil
}

// PuppyGet получает информацию о щенке.
func (s *ServiceImpl) PuppyGet(idPuppy string) (*domain.Puppy, *domain.Dog, *domain.Dog, error) {
	cacheKeyPuppy := fmt.Sprintf("puppy:%s", idPuppy)
	cacheKeyMother := fmt.Sprintf("dog:mother:%s", idPuppy)
	cacheKeyFather := fmt.Sprintf("dog:father:%s", idPuppy)

	cachedPuppy, err := s.Repository.RedisRepository.GetPuppy(cacheKeyPuppy)
	if err == nil && cachedPuppy != nil {
		cachedMother, err := s.Repository.RedisRepository.GetDog(cacheKeyMother)
		if err == nil && cachedMother != nil {
			cachedFather, err := s.Repository.RedisRepository.GetDog(cacheKeyFather)
			if err == nil && cachedFather != nil {
				return cachedPuppy, cachedMother, cachedFather, nil
			}
		}
	}

	puppy, err := s.Repository.PostgresRepository.PuppyGet(idPuppy)
	if err != nil {
		return nil, nil, nil, err
	}
	mother, err := s.Repository.PostgresRepository.DogGet(strconv.Itoa(puppy.MotherID))
	if err != nil {
		return nil, nil, nil, err
	}
	father, err := s.Repository.PostgresRepository.DogGet(strconv.Itoa(puppy.FatherID))
	if err != nil {
		return nil, nil, nil, err
	}

	go func() {
		err := s.Repository.RedisRepository.SetPuppy(cacheKeyPuppy, puppy)
		if err != nil {
			s.Logger.Error("Ошибка кеширования щенка", zap.Error(err))
		}

		err = s.Repository.RedisRepository.SetDog(cacheKeyMother, mother)
		if err != nil {
			s.Logger.Error("Ошибка кеширования матери щенка", zap.Error(err))
		}

		err = s.Repository.RedisRepository.SetDog(cacheKeyFather, father)
		if err != nil {
			s.Logger.Error("Ошибка кеширования отца щенка", zap.Error(err))
		}
	}()

	return puppy, mother, father, nil
}

// PuppyDelete удаляет информацию о щенке.
func (s *ServiceImpl) PuppyDelete(puppyID string) error {
	s.Repository.RedisRepository.FlushAll()
	imgUrls, err := s.Repository.PuppyDelete(puppyID)
	if err != nil {
		return err
	}
	// Удаляем файлы из S3
	for _, imgUrl := range imgUrls {
		err = s.Repository.S3Repository.DeleteFromS3(imgUrl)
		if err != nil {
			return fmt.Errorf("unable to delete from image URLs: %w", err)
		}
	}
	return err
}

// PuppyChangeArchived меняет состояние архива щенка.
func (s *ServiceImpl) PuppyChangeArchived(puppyID, archived, city, phone string) error {
	s.Repository.RedisRepository.FlushAll()
	return s.Repository.PuppyChangeArchived(puppyID, archived, city, phone)
}

// PuppyAdd добавляет информацию о щенке.
func (s *ServiceImpl) PuppyAdd(puppy *domain.Puppy, fileHeaders []*multipart.FileHeader) error {
	s.Repository.RedisRepository.FlushAll()
	urls, err := s.Repository.PutInS3(puppy.Name, fileHeaders, 3.0/2.0, 1200, 800)
	if err != nil {
		return err
	}
	puppy.Urls = urls
	println(puppy.Urls)
	return s.Repository.PuppyAdd(puppy)
}

// PuppyUpdate обновляет информацию о щенке.
func (s *ServiceImpl) PuppyUpdate(puppy *domain.Puppy, fileHeaders []*multipart.FileHeader) error {
	s.Repository.RedisRepository.FlushAll()
	newUrls, err := s.Repository.PutInS3(puppy.Name, fileHeaders, 3.0/2.0, 1200, 800)
	if err != nil {
		return err
	}
	// Объединение старых и новых URL-адресов
	puppy.Urls = append(puppy.Urls, newUrls...)

	currentUrlSet, newUrlSet, err := s.Repository.PuppyUpdate(puppy)
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

func (s *ServiceImpl) GetPagedPuppies(puppies []domain.Puppy, currentPage, perPage int) ([]domain.Puppy, int, error) {
	// Общее количество щенков
	puppiesCount := len(puppies)

	// Рассчитываем общее количество страниц
	totalPages := puppiesCount / perPage
	if puppiesCount%perPage != 0 {
		totalPages++
	}

	// Проверка текущей страницы на выход за пределы
	if currentPage > totalPages {
		currentPage = totalPages
	}
	if currentPage < 1 {
		currentPage = 1
	}

	// Вычисляем начальный и конечный индекс для текущей страницы
	startIndex := (currentPage - 1) * perPage
	endIndex := startIndex + perPage

	// Проверка на выход за границы
	if startIndex > puppiesCount {
		startIndex = puppiesCount
	}
	if endIndex > puppiesCount {
		endIndex = puppiesCount
	}

	log.Println(startIndex, "->", endIndex)

	// Получаем щенков для текущей страницы
	pagedPuppies := puppies[startIndex:endIndex]

	return pagedPuppies, totalPages, nil
}
