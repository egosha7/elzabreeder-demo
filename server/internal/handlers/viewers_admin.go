package handlers

import (
	"github.com/Masterminds/sprig/v3"
	"github.com/egosha7/site-go/internal/authMiddleware"
	"github.com/egosha7/site-go/internal/domain"
	"go.uber.org/zap"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

func (h *Handler) AdminPuppiesHandler(w http.ResponseWriter, r *http.Request, archived bool) {
	// Получение идентификатора пользователя из контекста
	userIDValue := r.Context().Value(authMiddleware.UserCtxKey)
	userID, ok := userIDValue.(int)
	if !ok {
		h.logger.Error("Ошибка получения идентификатора пользователя из контекста")
	} else {
		// Использование zap для логирования целого числа
		h.logger.Info("User ID", zap.Int("userID", userID))
	}
	lastUrlQuery := r.URL.RequestURI
	println(lastUrlQuery)
	// Валидация параметра страницы
	pageStr := r.URL.Query().Get("page")
	page, err := ValidatePage(pageStr)
	if err != nil {
		h.logger.Error("Ошибка при обработке параметра страницы", zap.Error(err))
		http.Error(w, "Ошибка при обработке параметра страницы", http.StatusBadRequest)
		page = 1 // если параметр страницы отсутствует или некорректен, установить его в 1
		return
	}

	// Валидация параметров поиска
	chocolates := r.URL.Query()["chocolate"]
	validatedChocolates, err := ValidateChocolates(chocolates)
	if err != nil {
		h.logger.Error("Ошибка при обработке параметра chocolate", zap.Error(err))
		http.Error(w, "Ошибка при обработке параметра chocolate", http.StatusBadRequest)
		return
	}

	genders := r.URL.Query()["gender"]
	validatedGenders, err := ValidateGender(genders)
	if err != nil {
		h.logger.Error("Ошибка при обработке параметра gender", zap.Error(err))
		http.Error(w, "Ошибка при обработке параметра gender", http.StatusBadRequest)
		return
	}

	readyToMoveStr := r.URL.Query().Get("readyToMove")
	readyToMove, err := ValidateReadyToMove(readyToMoveStr)
	if err != nil {
		h.logger.Error("Ошибка при обработке параметра readyToMove", zap.Error(err))
		http.Error(w, "Ошибка при обработке параметра readyToMove", http.StatusBadRequest)
		return
	}

	var idPuppy string
	var idParent string

	archivedParents := false

	getParams := contains(chocolates, genders, readyToMove)

	pagedPuppies, puppyReviews, totalPages, err := h.Services.PuppiesGet(
		validatedChocolates, validatedGenders, idPuppy, readyToMove, page, archived,
	)
	if err != nil {
		h.logger.Error("Ошибка при получении данных о щенках", zap.Error(err))
		http.Error(w, "Ошибка при получении данных о щенках", http.StatusInternalServerError)
		return
	}
	parentsList, err := h.Services.DogsGet(validatedChocolates, validatedGenders, idParent, archivedParents)
	if err != nil {
		h.logger.Error("Ошибка при получении данных о родителях", zap.Error(err))
		http.Error(w, "Ошибка при получении данных о родителях", http.StatusInternalServerError)
		return
	}

	if archived {
		t := template.Must(
			template.New("adminPuppyArchive").Funcs(sprig.FuncMap()).ParseFiles(
				"cmd/templates/admin/admin_menu_archive.html",
				"cmd/templates/parts/preloader.html",
				"cmd/templates/admin/admin_nav.html",
				"cmd/templates/admin/admin_footer.html",
			),
		)

		err = t.ExecuteTemplate(
			w, "adminPuppyArchive", struct {
				SelectedColors      []string
				SelectedGenders     []string
				IsReadyToMove       string
				GetParams           string
				Puppies             []domain.Puppy
				PuppiesWithFeedback map[int]int
				TotalPages          int
				CurrentPage         int
				Parents             []domain.Dog
			}{
				SelectedColors:      chocolates,
				SelectedGenders:     genders,
				IsReadyToMove:       readyToMove,
				GetParams:           getParams,
				Puppies:             pagedPuppies,
				PuppiesWithFeedback: puppyReviews,
				TotalPages:          totalPages,
				CurrentPage:         page,
				Parents:             parentsList,
			},
		)
		if err != nil {
			h.logger.Error("Ошибка вывода страницы с щенками (Архив)", zap.Error(err))
		}
	} else {
		t := template.Must(
			template.New("adminPuppyMenu").Funcs(sprig.FuncMap()).ParseFiles(
				"cmd/templates/admin/admin_menu_puppy.html",
				"cmd/templates/parts/preloader.html",
				"cmd/templates/admin/admin_nav.html",
				"cmd/templates/admin/admin_footer.html",
			),
		)

		err = t.ExecuteTemplate(
			w, "adminPuppyMenu", struct {
				SelectedColors  []string
				SelectedGenders []string
				IsReadyToMove   string
				GetParams       string
				Puppies         []domain.Puppy
				TotalPages      int
				CurrentPage     int
				Parents         []domain.Dog
			}{
				SelectedColors:  chocolates,
				SelectedGenders: genders,
				IsReadyToMove:   readyToMove,
				GetParams:       getParams,
				Puppies:         pagedPuppies,
				TotalPages:      totalPages,
				CurrentPage:     page,
				Parents:         parentsList,
			},
		)
		if err != nil {
			h.logger.Error("Ошибка вывода страницы с щенками", zap.Error(err))
		}
	}
}

func (h *Handler) AdminDogsHandler(w http.ResponseWriter, r *http.Request, archived bool) {
	// Получение идентификатора пользователя из контекста
	userIDValue := r.Context().Value(authMiddleware.UserCtxKey)
	userID, ok := userIDValue.(int)
	if !ok {
		h.logger.Error("Ошибка получения идентификатора пользователя из контекста")
	} else {
		// Использование zap для логирования целого числа
		h.logger.Info("User ID", zap.Int("userID", userID))
	}
	lastUrlQuery := r.URL.RequestURI
	println(lastUrlQuery)
	// Валидация параметра страницы
	pageStr := r.URL.Query().Get("page")
	page, err := ValidatePage(pageStr)
	if err != nil {
		h.logger.Error("Ошибка при обработке параметра страницы", zap.Error(err))
		http.Error(w, "Ошибка при обработке параметра страницы", http.StatusBadRequest)
		page = 1 // если параметр страницы отсутствует или некорректен, установить его в 1
		return
	}

	// Валидация параметров поиска
	chocolates := r.URL.Query()["chocolate"]
	validatedChocolates, err := ValidateChocolates(chocolates)
	if err != nil {
		h.logger.Error("Ошибка при обработке параметра chocolate", zap.Error(err))
		http.Error(w, "Ошибка при обработке параметра chocolate", http.StatusBadRequest)
		return
	}

	genders := r.URL.Query()["gender"]
	validatedGenders, err := ValidateGender(genders)
	if err != nil {
		h.logger.Error("Ошибка при обработке параметра gender", zap.Error(err))
		http.Error(w, "Ошибка при обработке параметра gender", http.StatusBadRequest)
		return
	}

	readyToMoveStr := r.URL.Query().Get("readyToMove")
	readyToMove, err := ValidateReadyToMove(readyToMoveStr)
	if err != nil {
		h.logger.Error("Ошибка при обработке параметра readyToMove", zap.Error(err))
		http.Error(w, "Ошибка при обработке параметра readyToMove", http.StatusBadRequest)
		return
	}

	var idParent string

	parentsList, err := h.Services.DogsGet(validatedChocolates, validatedGenders, idParent, archived)

	if err != nil {
		h.logger.Error("Ошибка при получении данных о собаках", zap.Error(err))
		http.Error(w, "Ошибка при получении данных о собаках", http.StatusInternalServerError)
		return
	}

	getParams := contains(chocolates, genders, readyToMove)
	log.Println(getParams)

	pagedDogs, totalPages, err := getPagedDogs(parentsList, page, 4)

	if archived {
		t := template.Must(
			template.New("adminMenuArchiveDog").Funcs(sprig.FuncMap()).ParseFiles(
				"cmd/templates/admin/admin_menu_archive_dog.html",
				"cmd/templates/parts/preloader.html",
				"cmd/templates/admin/admin_nav.html",
				"cmd/templates/admin/admin_footer.html",
			),
		)

		err = t.ExecuteTemplate(
			w, "adminMenuArchiveDog", struct {
				SelectedColors  []string
				SelectedGenders []string
				IsReadyToMove   string
				GetParams       string
				TotalPages      int
				CurrentPage     int
				Parents         []domain.Dog
			}{
				SelectedColors:  chocolates,
				SelectedGenders: genders,
				IsReadyToMove:   readyToMove,
				GetParams:       getParams,
				TotalPages:      totalPages,
				CurrentPage:     page,
				Parents:         pagedDogs,
			},
		)
		if err != nil {
			h.logger.Error("Ошибка вывода страницы с щенками (Архив)", zap.Error(err))
		}
	} else {
		t := template.Must(
			template.New("adminDogMenu").Funcs(sprig.FuncMap()).ParseFiles(
				"cmd/templates/admin/admin_menu_dog.html",
				"cmd/templates/parts/preloader.html",
				"cmd/templates/admin/admin_nav.html",
				"cmd/templates/admin/admin_footer.html",
			),
		)

		err = t.ExecuteTemplate(
			w, "adminDogMenu", struct {
				SelectedColors  []string
				SelectedGenders []string
				IsReadyToMove   string
				GetParams       string
				TotalPages      int
				CurrentPage     int
				Parents         []domain.Dog
			}{
				SelectedColors:  chocolates,
				SelectedGenders: genders,
				IsReadyToMove:   readyToMove,
				GetParams:       getParams,
				TotalPages:      totalPages,
				CurrentPage:     page,
				Parents:         pagedDogs,
			},
		)
		if err != nil {
			h.logger.Error("Ошибка вывода страницы с щенками", zap.Error(err))
		}
	}
}

// AdminReviewsView обрабатывает запрос на отображение страницы с отзывами.
func (h *Handler) AdminReviewsView(w http.ResponseWriter, r *http.Request, checked bool) {
	lastUrlQuery := r.URL.RequestURI
	println(lastUrlQuery)
	// Получаем параметр страницы из URL
	pageStr := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		h.logger.Error("Ошибка при обработке параметра страницы", zap.Error(err))
		page = 1 // если параметр страницы отсутствует или некорректен, установить его в 1
	}
	if page == 0 {
		page++
	}

	var idFeedback string
	reviews, puppyNames, err := h.Services.ReviewsGet(idFeedback, checked)
	pagedReviews, totalPages, err := getPagedReviews(reviews, page, 2)
	if err != nil {
		h.logger.Error("Ошибка при получении отзывов", zap.Error(err))
	}

	if checked {
		t := template.Must(
			template.New("adminReviews").Funcs(sprig.FuncMap()).ParseFiles(
				"cmd/templates/admin/admin_reviews.html",
				"cmd/templates/admin/admin_nav.html",
				"cmd/templates/admin/admin_footer.html",
				"cmd/templates/parts/preloader.html",
			),
		)

		err = t.ExecuteTemplate(
			w, "adminReviews", struct {
				Reviews               []domain.Feedback
				FeedbackWithPuppyName map[int]string
				TotalPages            int
				CurrentPage           int
			}{
				Reviews:               pagedReviews,
				FeedbackWithPuppyName: puppyNames,
				TotalPages:            totalPages,
				CurrentPage:           page,
			},
		)
		if err != nil {
			h.logger.Error("Ошибка вывода страницы с отзывами", zap.Error(err))
		}
	} else {
		t := template.Must(
			template.New("adminReviewsArchive").Funcs(sprig.FuncMap()).ParseFiles(
				"cmd/templates/admin/admin_reviews_archive.html",
				"cmd/templates/admin/admin_nav.html",
				"cmd/templates/admin/admin_footer.html",
				"cmd/templates/parts/preloader.html",
			),
		)

		err = t.ExecuteTemplate(
			w, "adminReviewsArchive", struct {
				Reviews               []domain.Feedback
				FeedbackWithPuppyName map[int]string
				TotalPages            int
				CurrentPage           int
			}{
				Reviews:               pagedReviews,
				FeedbackWithPuppyName: puppyNames,
				TotalPages:            totalPages,
				CurrentPage:           page,
			},
		)
		if err != nil {
			h.logger.Error("Ошибка вывода страницы с отзывами", zap.Error(err))
		}
	}
}
