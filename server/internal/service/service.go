package service

import (
	"github.com/egosha7/site-go/internal/domain"
	"github.com/egosha7/site-go/internal/repository"
	"go.uber.org/zap"
	"mime/multipart"
)

// Services представляет сервис для работы с пользователями.
//
//go:generate mockgen -source=service.go -destination=mock_service/mock.go
type Services interface {
	AddEmail(email string) error
	PuppiesGet(chocolates, genders []string, idPuppy, readyToMove string, page int, archived bool) ([]domain.Puppy, map[int]int, int, error)
	PuppyGet(idPuppy string) (*domain.Puppy, *domain.Dog, *domain.Dog, error)
	DogsGet(chocolates, genders []string, id string, archived bool) ([]domain.Dog, error)
	PuppyUpdate(puppy *domain.Puppy, fileHeaders []*multipart.FileHeader) error
	PuppyAdd(puppy *domain.Puppy, fileHeaders []*multipart.FileHeader) error
	PuppyDelete(puppyID string) error
	PuppyChangeArchived(puppyID, archived, city, phone string) error
	GetPagedPuppies(puppies []domain.Puppy, currentPage, perPage int) ([]domain.Puppy, int, error)
	DogChangeArchived(puppyID string, archived string) error
	DogAdd(puppy *domain.Dog, fileHeaders []*multipart.FileHeader) error
	DogUpdate(dog *domain.Dog, fileHeaders []*multipart.FileHeader) error
	ReviewsGet(idReview string, checked bool) ([]domain.Feedback, map[int]string, error)
	FeedbackGet(idPuppy, verified string) (*domain.Feedback, error)
	FeedbackAdd(feedback *domain.Feedback, fileHeaders []*multipart.FileHeader) error
	FeedbackUpdate(feedback *domain.Feedback, fileHeaders []*multipart.FileHeader) error
	FeedbackDelete(feedbackID string) error
	FeedbackChangeChecked(feedbackID string, checked string) error
}

type AuthorizationServices interface {
	GenerateToken(user *domain.User, ip string) (string, error)
	CreateUser(login, password string) error
	HashPassword(password string) (string, error)
	CheckPasswordHash(password, hash string) bool
}

type Service struct {
	Services
	AuthorizationServices
}

// ServiceImpl представляет реализацию Service.
type ServiceImpl struct {
	Repository *repository.Repository
	Logger     *zap.Logger
}

// NewUserService создает новый экземпляр UserService.
func NewUserService(repository *repository.Repository, logger *zap.Logger) *Service {
	return &Service{
		Services: &ServiceImpl{
			Repository: repository,
			Logger:     logger,
		},
		AuthorizationServices: &ServiceImpl{
			Repository: repository,
			Logger:     logger,
		},
	}
}
