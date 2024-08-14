package handlers_test

import (
	"errors"
	"fmt"
	"github.com/egosha7/site-go/internal/domain"
	"github.com/egosha7/site-go/internal/handlers"
	"github.com/egosha7/site-go/internal/logger"
	"github.com/egosha7/site-go/internal/service"
	"github.com/egosha7/site-go/internal/service/mock_service"
	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHandler_PuppiesView(t *testing.T) {
	type mockBehavior func(s *mock_service.MockServices, chocolates, genders []string, idPuppy, readyToMove string, archived bool, expectedPuppies []domain.Puppy, expectedReviews map[int]int, totalpages, page int)

	tests := []struct {
		name            string
		url             string
		chocolates      []string
		genders         []string
		idPuppy         string
		readyToMove     string
		archived        bool
		page            int
		totalpages      int
		expectedPuppies []domain.Puppy
		expectedReviews map[int]int
		mockBehavior    mockBehavior
		expectedCode    int
		expectedError   error
		expectedBody    string
	}{
		{
			name:        "Correct 200",
			url:         "/puppy",
			chocolates:  []string{},
			genders:     []string{},
			readyToMove: "",
			archived:    false,
			page:        1,
			totalpages:  1,
			expectedPuppies: []domain.Puppy{
				{
					ID:        1,
					Name:      "PuppyTestGo",
					Title:     "Example Puppy",
					Sex:       "Кобель",
					Price:     "30000",
					ReadyOut:  true,
					Archived:  false,
					City:      "Уфа",
					MotherID:  1,
					FatherID:  2,
					DateBirth: "20.08.2002",
					Color:     "Черный",
					Urls:      []string{"http://Puppy.com", "http://Puppy2.com", "http://Puppy3.com"},
				}, {
					ID:        2,
					Name:      "PuppyTestArchivedTrue",
					Title:     "Archived True",
					Sex:       "Кобель",
					Price:     "30000",
					ReadyOut:  true,
					Archived:  true,
					City:      "Уфа",
					MotherID:  1,
					FatherID:  2,
					DateBirth: "20.08.2002",
					Color:     "Черный",
					Urls:      []string{"http://Puppy.com", "http://Puppy2.com", "http://Puppy3.com"},
				},
			},
			expectedReviews: map[int]int{},
			mockBehavior: func(s *mock_service.MockServices, chocolates, genders []string, idPuppy, readyToMove string, archived bool, expectedPuppies []domain.Puppy, expectedReviews map[int]int, page, totalpages int) {
				s.EXPECT().PuppiesGet(chocolates, genders, idPuppy, readyToMove, page, archived).Return(
					expectedPuppies, expectedReviews, totalpages, nil,
				)
			},
			expectedCode: http.StatusOK,
		}, {
			name:        "Correct 200 With feedback",
			url:         "/archive",
			chocolates:  []string{},
			genders:     []string{},
			readyToMove: "",
			archived:    true,
			page:        1,
			totalpages:  1,
			expectedPuppies: []domain.Puppy{
				{
					ID:        1,
					Name:      "PuppyTestGo",
					Title:     "Example Puppy",
					Sex:       "Кобель",
					Price:     "30000",
					ReadyOut:  true,
					Archived:  true,
					City:      "Уфа",
					MotherID:  1,
					FatherID:  2,
					DateBirth: "20.08.2002",
					Color:     "Черный",
					Urls:      []string{"http://PuppyTestGo.com", "http://PuppyTestGo2.com", "http://PuppyTestGo3.com"},
				}, {
					ID:        2,
					Name:      "PuppyTestTrue",
					Title:     "PuppyTestTrue",
					Sex:       "Кобель",
					Price:     "30000",
					ReadyOut:  true,
					Archived:  true,
					City:      "Уфа",
					MotherID:  1,
					FatherID:  2,
					DateBirth: "20.08.2002",
					Color:     "Черный",
					Urls:      []string{"PuppyTestTrue", "PuppyTestTrue2", "PuppyTestTrue3"},
				},
			},
			expectedReviews: map[int]int{
				1: 23,
				3: 45,
				4: 67,
			},
			mockBehavior: func(s *mock_service.MockServices, chocolates, genders []string, idPuppy, readyToMove string, archived bool, expectedPuppies []domain.Puppy, expectedReviews map[int]int, totalpages, page int) {
				s.EXPECT().PuppiesGet(chocolates, genders, idPuppy, readyToMove, page, archived).Return(
					expectedPuppies, expectedReviews, totalpages, nil,
				)
			},
			expectedCode: http.StatusOK,
		}, {
			name:            "Failure service 500",
			url:             "/archive",
			expectedPuppies: []domain.Puppy{},
			expectedReviews: map[int]int{},
			totalpages:      1,
			mockBehavior: func(s *mock_service.MockServices, chocolates, genders []string, idPuppy, readyToMove string, archived bool, expectedPuppies []domain.Puppy, expectedReviews map[int]int, totalpages, page int) {
				s.EXPECT().PuppiesGet(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Return(expectedPuppies, expectedReviews, totalpages, errors.New("service error"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "Ошибка при получении данных о щенках",
		}, {
			name: "Failure validate page 400",
			url:  "/puppy?page=abc",
			mockBehavior: func(s *mock_service.MockServices, chocolates, genders []string, idPuppy, readyToMove string, archived bool, expectedPuppies []domain.Puppy, expectedReviews map[int]int, totalpages, page int) {
				s.EXPECT().PuppiesGet(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Times(0)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "Ошибка при обработке параметра страницы",
		}, {
			name: "Failure validate chocolates 400",
			url:  "/puppy?chocolate=<script>alert(XSS)</script>&chocolate=Бивер",
			mockBehavior: func(s *mock_service.MockServices, chocolates, genders []string, idPuppy, readyToMove string, archived bool, expectedPuppies []domain.Puppy, expectedReviews map[int]int, totalpages, page int) {
				// Ожидаем, что метод PuppiesGet не будет вызываться
				s.EXPECT().PuppiesGet(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Times(0)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "Ошибка при обработке параметра chocolate",
		}, {
			name: "Failure validate gender 400",
			url:  "/puppy?gender=SELECT",
			mockBehavior: func(s *mock_service.MockServices, chocolates, genders []string, idPuppy, readyToMove string, archived bool, expectedPuppies []domain.Puppy, expectedReviews map[int]int, totalpages, page int) {
				// Ожидаем, что метод PuppiesGet не будет вызываться
				s.EXPECT().PuppiesGet(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Times(0)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "Ошибка при обработке параметра gender",
		}, {
			name: "Failure validate readyToMove 400",
			url:  "/puppy?readyToMove=Да",
			mockBehavior: func(s *mock_service.MockServices, chocolates, genders []string, idPuppy, readyToMove string, archived bool, expectedPuppies []domain.Puppy, expectedReviews map[int]int, totalpages, page int) {
				// Ожидаем, что метод PuppiesGet не будет вызываться
				s.EXPECT().PuppiesGet(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Times(0)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "Ошибка при обработке параметра readyToMove",
		},
	}
	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				mockServices := mock_service.NewMockServices(ctrl)
				test.mockBehavior(
					mockServices, test.chocolates, test.genders, test.idPuppy, test.readyToMove, test.archived,
					test.expectedPuppies, test.expectedReviews, test.totalpages, test.page,
				)

				services := &service.Service{Services: mockServices}
				logger2, err := logger.SetupLogger()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Ошибка создания логгера: %v\n", err)
				}

				handler := handlers.NewHandler(services, logger2)

				router := chi.NewRouter()
				router.Get(
					"/puppy", func(w http.ResponseWriter, r *http.Request) {
						handler.PuppiesView(w, r, test.archived)
					},
				)
				router.Get(
					"/archive", func(w http.ResponseWriter, r *http.Request) {
						handler.PuppiesView(w, r, test.archived)
					},
				)

				req, err := http.NewRequest("GET", test.url, nil)
				assert.NoError(t, err)

				rr := httptest.NewRecorder()
				router.ServeHTTP(rr, req)

				assert.Equal(t, test.expectedCode, rr.Code)

				body := rr.Body.String()
				// Выводим тело ответа для отладки
				if rr.Code != test.expectedCode {
					fmt.Printf("Response body: %s\n", body)
				}

				assert.Contains(t, body, test.expectedBody)

				for _, puppy := range test.expectedPuppies {
					assert.Contains(t, body, puppy.Name)
					assert.Contains(t, body, puppy.Title)
					for _, url := range puppy.Urls {
						assert.Contains(t, body, url)
					}
				}

				if len(test.expectedReviews) > 1 {
					assert.Contains(t, body, "Есть отзыв")
				}
			},
		)
	}
}

func TestHandler_PuppyView(t *testing.T) {
	type mockBehavior func(s *mock_service.MockServices, inputIdPuppy string, inputVerify string, expectedPuppy *domain.Puppy, expectedMother *domain.Dog, expectedFather *domain.Dog)

	tests := []struct {
		name             string
		inputIdPuppy     string
		inputVerify      string
		expectedPuppy    *domain.Puppy
		expectedMother   *domain.Dog
		expectedFather   *domain.Dog
		expectedFeedback *domain.Feedback
		setup            func(h *handlers.Handler)
		mockBehavior     mockBehavior
		expectedCode     int
		expectedError    error
		expectedBody     string
	}{
		{
			name:         "Correct 200",
			inputIdPuppy: "12",
			inputVerify:  "true",
			expectedPuppy: &domain.Puppy{
				ID:        12,
				Name:      "PuppyTestGo",
				Title:     "Example Puppy",
				Sex:       "Кобель",
				Price:     "30000",
				ReadyOut:  true,
				Archived:  false,
				City:      "Уфа",
				MotherID:  1,
				FatherID:  2,
				DateBirth: "20.08.2002",
				Color:     "Черный",
				Urls:      []string{"http://Puppy.com", "http://Puppy2.com", "http://Puppy3.com"},
			},
			expectedMother: &domain.Dog{
				ID:       1,
				Name:     "Mother",
				Title:    "Example Mother",
				Gender:   "Сука",
				Color:    "Черный",
				Archived: false,
				Urls:     []string{"http://Mother.com", "http://Mother2.com", "http://Mother3.com"},
			},
			expectedFather: &domain.Dog{
				ID:       1,
				Name:     "Father",
				Title:    "Example Father",
				Gender:   "Кобель",
				Color:    "Черный",
				Archived: false,
				Urls:     []string{"http://Father.com", "http://Father2.com", "http://Father3.com"},
			},
			mockBehavior: func(s *mock_service.MockServices, inputIdPuppy string, inputVerify string, expectedPuppy *domain.Puppy, expectedMother *domain.Dog, expectedFather *domain.Dog) {
				s.EXPECT().PuppyGet(inputIdPuppy).Return(
					expectedPuppy, expectedMother, expectedFather, nil,
				)
				s.EXPECT().FeedbackGet(inputIdPuppy, inputVerify).Return(
					&domain.Feedback{}, nil,
				)
			},
			expectedCode: http.StatusOK,
		}, {
			name:           "Not Found 404 (Puppy ID > 1000)",
			inputIdPuppy:   "1234",
			inputVerify:    "true",
			expectedPuppy:  &domain.Puppy{},
			expectedMother: &domain.Dog{},
			expectedFather: &domain.Dog{},
			mockBehavior: func(s *mock_service.MockServices, inputIdPuppy string, inputVerify string, expectedPuppy *domain.Puppy, expectedMother *domain.Dog, expectedFather *domain.Dog) {
				s.EXPECT().PuppyGet(
					gomock.Any(),
				).Times(0)
				s.EXPECT().FeedbackGet(
					gomock.Any(), gomock.Any(),
				).Times(0)
			},
			expectedCode: http.StatusNotFound,
		}, {
			name:           "Not Found 404 (Puppy ID with string)",
			inputIdPuppy:   "abc",
			inputVerify:    "true",
			expectedPuppy:  &domain.Puppy{},
			expectedMother: &domain.Dog{},
			expectedFather: &domain.Dog{},
			mockBehavior: func(s *mock_service.MockServices, inputIdPuppy string, inputVerify string, expectedPuppy *domain.Puppy, expectedMother *domain.Dog, expectedFather *domain.Dog) {
				s.EXPECT().PuppyGet(
					gomock.Any(),
				).Times(0)
				s.EXPECT().FeedbackGet(
					gomock.Any(), gomock.Any(),
				).Times(0)
			},
			expectedCode: http.StatusNotFound,
		}, {
			name:           "Bad Request 500 (Service PuppyGet failure)",
			inputIdPuppy:   "123",
			inputVerify:    "true",
			expectedPuppy:  &domain.Puppy{},
			expectedMother: &domain.Dog{},
			expectedFather: &domain.Dog{},
			mockBehavior: func(s *mock_service.MockServices, inputIdPuppy string, inputVerify string, expectedPuppy *domain.Puppy, expectedMother *domain.Dog, expectedFather *domain.Dog) {
				s.EXPECT().PuppyGet(inputIdPuppy).Return(nil, nil, nil, errors.New("service error"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "Ошибка сервиса: не удалось получить информацию о щенке",
		}, {
			name:           "Bad Request 500 (Service FeedbackGet failure)",
			inputIdPuppy:   "123",
			inputVerify:    "true",
			expectedPuppy:  &domain.Puppy{},
			expectedMother: &domain.Dog{},
			expectedFather: &domain.Dog{},
			mockBehavior: func(s *mock_service.MockServices, inputIdPuppy string, inputVerify string, expectedPuppy *domain.Puppy, expectedMother *domain.Dog, expectedFather *domain.Dog) {
				s.EXPECT().PuppyGet(inputIdPuppy).Return(
					expectedPuppy, expectedMother, expectedFather, nil,
				)
				s.EXPECT().FeedbackGet(inputIdPuppy, inputVerify).Return(
					nil, errors.New("service error"),
				)
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "Ошибка сервиса: не удалось получить информацию об отзыве",
		}, {
			name:           "Bad Request 500 (Template execute failure)",
			inputIdPuppy:   "123",
			inputVerify:    "true",
			expectedPuppy:  &domain.Puppy{},
			expectedMother: &domain.Dog{},
			expectedFather: &domain.Dog{},
			setup: func(h *handlers.Handler) {
				h.ExecuteTemplate = func(t *template.Template, w http.ResponseWriter, name string, data interface{}) error {
					return errors.New("template execute error")
				}
			},
			mockBehavior: func(s *mock_service.MockServices, inputIdPuppy string, inputVerify string, expectedPuppy *domain.Puppy, expectedMother *domain.Dog, expectedFather *domain.Dog) {
				s.EXPECT().PuppyGet(inputIdPuppy).Return(expectedPuppy, expectedMother, expectedFather, nil)
				s.EXPECT().FeedbackGet(inputIdPuppy, inputVerify).Return(
					&domain.Feedback{}, nil,
				)
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "Ошибка сервера: не удалось отобразить страницу",
		},
	}
	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				mockServices := mock_service.NewMockServices(ctrl)
				test.mockBehavior(
					mockServices, test.inputIdPuppy, test.inputVerify, test.expectedPuppy, test.expectedMother,
					test.expectedFather,
				)

				services := &service.Service{Services: mockServices}
				logger2, err := logger.SetupLogger()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Ошибка создания логгера: %v\n", err)
				}

				handler := handlers.NewHandler(services, logger2)

				// Настройка перед каждым тестом
				if test.setup != nil {
					test.setup(handler)
				}

				router := chi.NewRouter()
				router.Get("/puppy/{id}", handler.PuppyView)

				req, err := http.NewRequest("GET", "/puppy/"+test.inputIdPuppy, nil)
				assert.NoError(t, err)

				rr := httptest.NewRecorder()
				router.ServeHTTP(rr, req)

				assert.Equal(t, test.expectedCode, rr.Code)

				body := rr.Body.String()
				// Выводим тело ответа для отладки
				if rr.Code != test.expectedCode {
					fmt.Printf("Response body: %s\n", body)
				}

				assert.Contains(t, body, test.expectedBody)

				assert.Contains(t, body, test.expectedPuppy.Name)
				assert.Contains(t, body, test.expectedPuppy.Title)
				for _, url := range test.expectedPuppy.Urls {
					assert.Contains(t, body, url)
				}

				assert.Contains(t, body, test.expectedMother.Name)
				assert.Contains(t, body, test.expectedMother.Title)
				for _, url := range test.expectedMother.Urls {
					assert.Contains(t, body, url)
				}

				assert.Contains(t, body, test.expectedFather.Name)
				assert.Contains(t, body, test.expectedFather.Title)
				for _, url := range test.expectedFather.Urls {
					assert.Contains(t, body, url)
				}
			},
		)
	}
}

func TestHandler_ReviewsView(t *testing.T) {
	type mockBehavior func(s *mock_service.MockServices, idReview string, checked bool, expectedReviews []domain.Feedback, expectedPuppyNames map[int]string)

	tests := []struct {
		name               string
		url                string
		page               string
		checked            bool
		idReview           string
		expectedReviews    []domain.Feedback
		expectedPuppyNames map[int]string
		mockBehavior       mockBehavior
		expectedCode       int
		expectedError      error
		expectedBody       string
	}{
		{
			name:    "Correct 200",
			url:     "/feedback",
			page:    "1",
			checked: true,
			expectedReviews: []domain.Feedback{
				{ID: 1, Name: "Review 1"},
				{ID: 2, Name: "Review 2"},
			},
			expectedPuppyNames: map[int]string{
				1: "Puppy 1",
				2: "Puppy 2",
			},
			mockBehavior: func(s *mock_service.MockServices, idReview string, checked bool, expectedReviews []domain.Feedback, expectedPuppyNames map[int]string) {
				s.EXPECT().ReviewsGet(idReview, checked).Return(expectedReviews, expectedPuppyNames, nil)
			},
			expectedCode: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				mockServices := mock_service.NewMockServices(ctrl)
				test.mockBehavior(
					mockServices, test.idReview, test.checked, test.expectedReviews, test.expectedPuppyNames,
				)

				services := &service.Service{Services: mockServices}
				logger2, err := logger.SetupLogger()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Ошибка создания логгера: %v\n", err)
				}

				handler := handlers.NewHandler(services, logger2)

				router := chi.NewRouter()
				router.Get("/feedback", handler.ReviewsView)

				req, err := http.NewRequest("GET", test.url, nil)
				assert.NoError(t, err)

				rr := httptest.NewRecorder()
				router.ServeHTTP(rr, req)

				assert.Equal(t, test.expectedCode, rr.Code)

				body := rr.Body.String()
				if rr.Code != test.expectedCode {
					fmt.Printf("Response body: %s\n", body)
				}

				assert.Contains(t, body, test.expectedBody)
				for _, review := range test.expectedReviews {
					assert.Contains(t, body, review.Name)
				}
			},
		)
	}
}

func TestHandler_MainView(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(h *handlers.Handler)
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Correct 200",
			expectedCode: http.StatusOK,
			expectedBody: "html",
		},
		{
			name: "Template failure parse 500",
			setup: func(h *handlers.Handler) {
				h.ParseTemplate = func(filenames ...string) (*template.Template, error) {
					return template.New(""), errors.New("template parse error")
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "Ошибка вывода главной страницы",
		}, {
			name: "Template failure execute 500",
			setup: func(h *handlers.Handler) {
				h.ParseTemplate = func(filenames ...string) (*template.Template, error) {
					return template.New("index"), nil
				}
				h.ExecuteTemplate = func(t *template.Template, w http.ResponseWriter, name string, data interface{}) error {
					return errors.New("template execute error")
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "Ошибка вывода главной страницы",
		},
	}

	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				mockServices := mock_service.NewMockServices(ctrl)
				services := &service.Service{Services: mockServices}
				logger2, err := logger.SetupLogger()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Ошибка создания логгера: %v\n", err)
				}

				handler := handlers.NewHandler(services, logger2)

				// Настройка перед каждым тестом
				if test.setup != nil {
					test.setup(handler)
				}

				router := chi.NewRouter()
				router.Get("/", handler.MainView)

				req, err := http.NewRequest("GET", "/", nil)
				assert.NoError(t, err)

				rr := httptest.NewRecorder()
				router.ServeHTTP(rr, req)

				assert.Equal(t, test.expectedCode, rr.Code)

				body := rr.Body.String()
				if rr.Code != test.expectedCode {
					fmt.Printf("Response body: %s\n", body)
				}

				assert.Contains(t, body, test.expectedBody)
			},
		)
	}
}

func TestHandler_AuthView(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(h *handlers.Handler)
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Correct 200",
			expectedCode: http.StatusOK,
			expectedBody: "html",
		},
		{
			name: "Template failure parse 500",
			setup: func(h *handlers.Handler) {
				h.ParseTemplate = func(filenames ...string) (*template.Template, error) {
					return template.New(""), errors.New("template parse error")
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "Ошибка вывода страницы авторизации",
		}, {
			name: "Template failure execute 500",
			setup: func(h *handlers.Handler) {
				h.ParseTemplate = func(filenames ...string) (*template.Template, error) {
					return template.New("index"), nil
				}
				h.ExecuteTemplate = func(t *template.Template, w http.ResponseWriter, name string, data interface{}) error {
					return errors.New("template execute error")
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "Ошибка вывода страницы авторизации",
		},
	}

	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				mockServices := mock_service.NewMockServices(ctrl)
				services := &service.Service{Services: mockServices}
				logger2, err := logger.SetupLogger()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Ошибка создания логгера: %v\n", err)
				}

				handler := handlers.NewHandler(services, logger2)

				// Настройка перед каждым тестом
				if test.setup != nil {
					test.setup(handler)
				}

				router := chi.NewRouter()
				router.Get("/auth", handler.AuthView)

				req, err := http.NewRequest("GET", "/auth", nil)
				assert.NoError(t, err)

				rr := httptest.NewRecorder()
				router.ServeHTTP(rr, req)

				assert.Equal(t, test.expectedCode, rr.Code)

				body := rr.Body.String()
				if rr.Code != test.expectedCode {
					fmt.Printf("Response body: %s\n", body)
				}

				assert.Contains(t, body, test.expectedBody)
			},
		)
	}
}

func TestHandler_NewFeedbackView(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(h *handlers.Handler)
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Correct 200",
			expectedCode: http.StatusOK,
			expectedBody: "html",
		}, {
			name: "Template failure execute 500",
			setup: func(h *handlers.Handler) {
				h.ExecuteTemplate = func(t *template.Template, w http.ResponseWriter, name string, data interface{}) error {
					return errors.New("template execute error")
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "Ошибка вывода страницы для добавления отзыва",
		},
	}

	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				mockServices := mock_service.NewMockServices(ctrl)
				services := &service.Service{Services: mockServices}
				logger2, err := logger.SetupLogger()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Ошибка создания логгера: %v\n", err)
				}

				handler := handlers.NewHandler(services, logger2)

				// Настройка перед каждым тестом
				if test.setup != nil {
					test.setup(handler)
				}

				router := chi.NewRouter()
				router.Get("/feedback/new", handler.NewFeedbackView)

				req, err := http.NewRequest("GET", "/feedback/new", nil)
				assert.NoError(t, err)

				rr := httptest.NewRecorder()
				router.ServeHTTP(rr, req)

				assert.Equal(t, test.expectedCode, rr.Code)

				body := rr.Body.String()
				if rr.Code != test.expectedCode {
					fmt.Printf("Response body: %s\n", body)
				}

				assert.Contains(t, body, test.expectedBody)
			},
		)
	}
}

func TestHandler_ArticleView(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(h *handlers.Handler)
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Correct 200",
			expectedCode: http.StatusOK,
			expectedBody: "html",
		}, {
			name: "Template failure execute 500",
			setup: func(h *handlers.Handler) {
				h.ExecuteTemplate = func(t *template.Template, w http.ResponseWriter, name string, data interface{}) error {
					return errors.New("template execute error")
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "Ошибка вывода страницы статьи",
		},
	}

	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				mockServices := mock_service.NewMockServices(ctrl)
				services := &service.Service{Services: mockServices}
				logger2, err := logger.SetupLogger()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Ошибка создания логгера: %v\n", err)
				}

				handler := handlers.NewHandler(services, logger2)

				// Настройка перед каждым тестом
				if test.setup != nil {
					test.setup(handler)
				}

				router := chi.NewRouter()
				router.Get("/article", handler.ArticleView)

				req, err := http.NewRequest("GET", "/article", nil)
				assert.NoError(t, err)

				rr := httptest.NewRecorder()
				router.ServeHTTP(rr, req)

				assert.Equal(t, test.expectedCode, rr.Code)

				body := rr.Body.String()
				if rr.Code != test.expectedCode {
					fmt.Printf("Response body: %s\n", body)
				}

				assert.Contains(t, body, test.expectedBody)
			},
		)
	}
}

func TestHandler_NotFoundView(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(h *handlers.Handler)
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Correct 404",
			expectedCode: http.StatusNotFound,
			expectedBody: "html",
		}, {
			name: "Template failure execute 500",
			setup: func(h *handlers.Handler) {
				h.ExecuteTemplate = func(t *template.Template, w http.ResponseWriter, name string, data interface{}) error {
					return errors.New("template execute error")
				}
			},
			expectedCode: http.StatusNotFound,
			expectedBody: "Ошибка вывода страницы для кода 404",
		},
	}

	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				mockServices := mock_service.NewMockServices(ctrl)
				services := &service.Service{Services: mockServices}
				logger2, err := logger.SetupLogger()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Ошибка создания логгера: %v\n", err)
				}

				handler := handlers.NewHandler(services, logger2)

				// Настройка перед каждым тестом
				if test.setup != nil {
					test.setup(handler)
				}

				router := chi.NewRouter()
				router.Get(
					"/test", func(w http.ResponseWriter, r *http.Request) {},
				)
				// Обработчик для 404
				router.NotFound(
					func(w http.ResponseWriter, r *http.Request) {
						handler.NotFoundView(w, r)
					},
				)

				req, err := http.NewRequest("GET", "/article", nil)
				assert.NoError(t, err)

				rr := httptest.NewRecorder()
				router.ServeHTTP(rr, req)

				assert.Equal(t, test.expectedCode, rr.Code)

				body := rr.Body.String()
				if rr.Code != test.expectedCode {
					fmt.Printf("Response body: %s\n", body)
				}

				assert.Contains(t, body, test.expectedBody)
			},
		)
	}
}
