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
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHandler_AdminPuppiesHandler(t *testing.T) {
	// Изменяем рабочую директорию на корень проекта
	err := os.Chdir("../../")
	if err != nil {
		t.Fatalf("не удалось изменить рабочую директорию: %v", err)
	}
	type mockBehavior func(s *mock_service.MockServices, chocolates, genders []string, idPuppy, readyToMove string, archived bool, expectedPuppies []domain.Puppy, expectedReviews map[int]int, totalpages, page int, idParent string, archivedParents bool, expectedParents []domain.Dog)

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
		idParent        string
		archivedParents bool
		expectedPuppies []domain.Puppy
		expectedReviews map[int]int
		expectedParents []domain.Dog
		mockBehavior    mockBehavior
		expectedCode    int
		expectedError   error
		expectedBody    string
	}{
		{
			name:            "Correct 200",
			url:             "/puppy",
			chocolates:      []string{},
			genders:         []string{},
			readyToMove:     "",
			archived:        false,
			page:            1,
			totalpages:      1,
			archivedParents: false,
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
			expectedParents: []domain.Dog{
				{
					ID:       1,
					Name:     "DogTestMale",
					Title:    "Male",
					Gender:   "Кобель",
					Color:    "Черный",
					Archived: false,
					Urls:     []string{"http://Dog.com", "http://Dog2.com", "http://Dog3.com"},
				}, {
					ID:       2,
					Name:     "DogTestGirl",
					Title:    "Girl",
					Gender:   "Сука",
					Color:    "Черный",
					Archived: false,
					Urls:     []string{"http://Dog.com", "http://Dog2.com", "http://Dog3.com"},
				},
			},
			mockBehavior: func(s *mock_service.MockServices, chocolates, genders []string, idPuppy, readyToMove string, archived bool, expectedPuppies []domain.Puppy, expectedReviews map[int]int, page, totalpages int, idParent string, archivedParents bool, expectedParents []domain.Dog) {
				s.EXPECT().PuppiesGet(chocolates, genders, idPuppy, readyToMove, page, archived).Return(
					expectedPuppies, expectedReviews, totalpages, nil,
				)
				s.EXPECT().DogsGet(chocolates, genders, idPuppy, archivedParents).Return(
					expectedParents, nil,
				)
			},
			expectedCode: http.StatusOK,
		}, {
			name:            "Correct 200 With feedback",
			url:             "/archive",
			chocolates:      []string{},
			genders:         []string{},
			readyToMove:     "",
			archived:        true,
			page:            1,
			totalpages:      1,
			archivedParents: false,
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
			expectedParents: []domain.Dog{
				{
					ID:       1,
					Name:     "DogTestMale",
					Title:    "Male",
					Gender:   "Кобель",
					Color:    "Черный",
					Archived: false,
					Urls:     []string{"http://Dog.com", "http://Dog2.com", "http://Dog3.com"},
				}, {
					ID:       2,
					Name:     "DogTestGirl",
					Title:    "Girl",
					Gender:   "Сука",
					Color:    "Черный",
					Archived: false,
					Urls:     []string{"http://Dog.com", "http://Dog2.com", "http://Dog3.com"},
				},
			},
			mockBehavior: func(s *mock_service.MockServices, chocolates, genders []string, idPuppy, readyToMove string, archived bool, expectedPuppies []domain.Puppy, expectedReviews map[int]int, totalpages, page int, idParent string, archivedParents bool, expectedParents []domain.Dog) {
				s.EXPECT().PuppiesGet(chocolates, genders, idPuppy, readyToMove, page, archived).Return(
					expectedPuppies, expectedReviews, totalpages, nil,
				)
				s.EXPECT().DogsGet(chocolates, genders, idPuppy, archivedParents).Return(
					expectedParents, nil,
				)
			},
			expectedCode: http.StatusOK,
		}, {
			name:            "Failure service PuppiesGet 500",
			url:             "/archive",
			expectedPuppies: []domain.Puppy{},
			expectedReviews: map[int]int{},
			totalpages:      1,
			mockBehavior: func(s *mock_service.MockServices, chocolates, genders []string, idPuppy, readyToMove string, archived bool, expectedPuppies []domain.Puppy, expectedReviews map[int]int, totalpages, page int, idParent string, archivedParents bool, expectedParents []domain.Dog) {
				s.EXPECT().PuppiesGet(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Return(expectedPuppies, expectedReviews, totalpages, errors.New("service error"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "Ошибка при получении данных о щенках",
		}, {
			name:            "Failure service DogsGet 500",
			url:             "/archive",
			expectedPuppies: []domain.Puppy{},
			expectedParents: []domain.Dog{},
			expectedReviews: map[int]int{},
			totalpages:      1,
			mockBehavior: func(s *mock_service.MockServices, chocolates, genders []string, idPuppy, readyToMove string, archived bool, expectedPuppies []domain.Puppy, expectedReviews map[int]int, totalpages, page int, idParent string, archivedParents bool, expectedParents []domain.Dog) {
				s.EXPECT().PuppiesGet(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Return(expectedPuppies, expectedReviews, totalpages, nil)
				s.EXPECT().DogsGet(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Return(expectedParents, errors.New("service error"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "Ошибка при получении данных о родителях",
		}, {
			name: "Failure validate page 400",
			url:  "/puppy?page=abc",
			mockBehavior: func(s *mock_service.MockServices, chocolates, genders []string, idPuppy, readyToMove string, archived bool, expectedPuppies []domain.Puppy, expectedReviews map[int]int, totalpages, page int, idParent string, archivedParents bool, expectedParents []domain.Dog) {
				s.EXPECT().PuppiesGet(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Times(0)
				s.EXPECT().DogsGet(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Times(0)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "Ошибка при обработке параметра страницы",
		}, {
			name: "Failure validate chocolates 400",
			url:  "/puppy?chocolate=<script>alert(XSS)</script>&chocolate=Бивер",
			mockBehavior: func(s *mock_service.MockServices, chocolates, genders []string, idPuppy, readyToMove string, archived bool, expectedPuppies []domain.Puppy, expectedReviews map[int]int, totalpages, page int, idParent string, archivedParents bool, expectedParents []domain.Dog) {
				// Ожидаем, что метод PuppiesGet не будет вызываться
				s.EXPECT().PuppiesGet(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Times(0)
				s.EXPECT().DogsGet(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Times(0)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "Ошибка при обработке параметра chocolate",
		}, {
			name: "Failure validate gender 400",
			url:  "/puppy?gender=SELECT",
			mockBehavior: func(s *mock_service.MockServices, chocolates, genders []string, idPuppy, readyToMove string, archived bool, expectedPuppies []domain.Puppy, expectedReviews map[int]int, totalpages, page int, idParent string, archivedParents bool, expectedParents []domain.Dog) {
				// Ожидаем, что метод PuppiesGet не будет вызываться
				s.EXPECT().PuppiesGet(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Times(0)
				s.EXPECT().DogsGet(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Times(0)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "Ошибка при обработке параметра gender",
		}, {
			name: "Failure validate readyToMove 400",
			url:  "/puppy?readyToMove=Да",
			mockBehavior: func(s *mock_service.MockServices, chocolates, genders []string, idPuppy, readyToMove string, archived bool, expectedPuppies []domain.Puppy, expectedReviews map[int]int, totalpages, page int, idParent string, archivedParents bool, expectedParents []domain.Dog) {
				// Ожидаем, что метод PuppiesGet не будет вызываться
				s.EXPECT().PuppiesGet(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Times(0)
				s.EXPECT().DogsGet(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
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
					test.expectedPuppies, test.expectedReviews, test.totalpages, test.page, test.idParent,
					test.archivedParents, test.expectedParents,
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
						handler.AdminPuppiesHandler(w, r, test.archived)
					},
				)
				router.Get(
					"/archive", func(w http.ResponseWriter, r *http.Request) {
						handler.AdminPuppiesHandler(w, r, test.archived)
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
			},
		)
	}
}

func TestHandler_AdminDogsHandler(t *testing.T) {
	type mockBehavior func(s *mock_service.MockServices, chocolates, genders []string, idDog, readyToMove string, archived bool, totalpages, page int, expectedDogs []domain.Dog)

	tests := []struct {
		name          string
		url           string
		chocolates    []string
		genders       []string
		idPuppy       string
		readyToMove   string
		archived      bool
		page          int
		totalpages    int
		idDog         string
		expectedDogs  []domain.Dog
		mockBehavior  mockBehavior
		expectedCode  int
		expectedError error
		expectedBody  string
	}{
		{
			name:        "Correct 200",
			url:         "/dogs",
			chocolates:  []string{},
			genders:     []string{},
			readyToMove: "",
			archived:    false,
			page:        1,
			totalpages:  1,
			expectedDogs: []domain.Dog{
				{
					ID:       1,
					Name:     "DogTestMale",
					Title:    "Male",
					Gender:   "Кобель",
					Color:    "Черный",
					Archived: false,
					Urls:     []string{"http://Dog.com", "http://Dog2.com", "http://Dog3.com"},
				}, {
					ID:       2,
					Name:     "DogTestGirl",
					Title:    "Girl",
					Gender:   "Сука",
					Color:    "Черный",
					Archived: false,
					Urls:     []string{"http://Dog.com", "http://Dog2.com", "http://Dog3.com"},
				},
			},
			mockBehavior: func(s *mock_service.MockServices, chocolates, genders []string, idDog, readyToMove string, archived bool, page, totalpages int, expectedDogs []domain.Dog) {
				s.EXPECT().DogsGet(chocolates, genders, idDog, archived).Return(
					expectedDogs, nil,
				)
			},
			expectedCode: http.StatusOK,
		}, {
			name:        "Correct 200",
			url:         "/archive/dogs",
			chocolates:  []string{},
			genders:     []string{},
			readyToMove: "",
			archived:    true,
			page:        1,
			totalpages:  1,
			expectedDogs: []domain.Dog{
				{
					ID:       1,
					Name:     "DogTestMale",
					Title:    "Male",
					Gender:   "Кобель",
					Color:    "Черный",
					Archived: false,
					Urls:     []string{"http://Dog.com", "http://Dog2.com", "http://Dog3.com"},
				}, {
					ID:       2,
					Name:     "DogTestGirl",
					Title:    "Girl",
					Gender:   "Сука",
					Color:    "Черный",
					Archived: false,
					Urls:     []string{"http://Dog.com", "http://Dog2.com", "http://Dog3.com"},
				},
			},
			mockBehavior: func(s *mock_service.MockServices, chocolates, genders []string, idDog, readyToMove string, archived bool, page, totalpages int, expectedDogs []domain.Dog) {
				s.EXPECT().DogsGet(chocolates, genders, idDog, archived).Return(
					expectedDogs, nil,
				)
			},
			expectedCode: http.StatusOK,
		}, {
			name:       "Failure service 500",
			url:        "/archive/dogs",
			totalpages: 1,
			mockBehavior: func(s *mock_service.MockServices, chocolates, genders []string, idDog, readyToMove string, archived bool, page, totalpages int, expectedParents []domain.Dog) {
				s.EXPECT().DogsGet(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Return(expectedParents, errors.New("service error"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "Ошибка при получении данных о собаках",
		}, {
			name: "Failure validate page 400",
			url:  "/dogs?page=abc",
			mockBehavior: func(s *mock_service.MockServices, chocolates, genders []string, idDog, readyToMove string, archived bool, page, totalpages int, expectedParents []domain.Dog) {
				s.EXPECT().DogsGet(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Times(0)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "Ошибка при обработке параметра страницы",
		}, {
			name: "Failure validate chocolates 400",
			url:  "/dogs?chocolate=<script>alert(XSS)</script>&chocolate=Бивер",
			mockBehavior: func(s *mock_service.MockServices, chocolates, genders []string, idDog, readyToMove string, archived bool, page, totalpages int, expectedParents []domain.Dog) {
				// Ожидаем, что метод DogsGet не будет вызываться
				s.EXPECT().DogsGet(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Times(0)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "Ошибка при обработке параметра chocolate",
		}, {
			name: "Failure validate gender 400",
			url:  "/dogs?gender=SELECT",
			mockBehavior: func(s *mock_service.MockServices, chocolates, genders []string, idDog, readyToMove string, archived bool, page, totalpages int, expectedParents []domain.Dog) {
				// Ожидаем, что метод DogsGet не будет вызываться
				s.EXPECT().DogsGet(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Times(0)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "Ошибка при обработке параметра gender",
		}, {
			name: "Failure validate readyToMove 400",
			url:  "/dogs?readyToMove=Да",
			mockBehavior: func(s *mock_service.MockServices, chocolates, genders []string, idDog, readyToMove string, archived bool, page, totalpages int, expectedParents []domain.Dog) {
				// Ожидаем, что метод DogsGet не будет вызываться
				s.EXPECT().DogsGet(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
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

				MockServices := mock_service.NewMockServices(ctrl)
				test.mockBehavior(
					MockServices, test.chocolates, test.genders, test.idDog, test.readyToMove, test.archived,
					test.page, test.totalpages, test.expectedDogs,
				)

				services := &service.Service{Services: MockServices}
				logger2, err := logger.SetupLogger()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Ошибка создания логгера: %v\n", err)
				}

				handler := handlers.NewHandler(services, logger2)

				router := chi.NewRouter()
				router.Get(
					"/dogs", func(w http.ResponseWriter, r *http.Request) {
						handler.AdminDogsHandler(w, r, test.archived)
					},
				)
				router.Get(
					"/archive/dogs", func(w http.ResponseWriter, r *http.Request) {
						handler.AdminDogsHandler(w, r, test.archived)
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

				if rr.Code == http.StatusOK {
					if test.archived == true {
						assert.Contains(t, body, "Взрослые собаки (Архив)")
					} else {
						assert.Contains(t, body, "Перейти к архиву")
					}
				}

				assert.Contains(t, body, test.expectedBody)

				for _, dog := range test.expectedDogs {
					assert.Contains(t, body, dog.Name)
					assert.Contains(t, body, dog.Title)
					for _, url := range dog.Urls {
						assert.Contains(t, body, url)
					}
				}
			},
		)
	}
}
