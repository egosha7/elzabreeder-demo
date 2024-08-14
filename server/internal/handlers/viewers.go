package handlers

import (
	"github.com/Masterminds/sprig/v3"
	"github.com/egosha7/site-go/internal/domain"
	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"html/template"
	"net/http"
	"strconv"
)

// MainView обрабатывает запрос на отображение главной страницы.
func (h *Handler) MainView(w http.ResponseWriter, r *http.Request) {
	t, err := h.ParseTemplate(
		"cmd/templates/index.html",
		"cmd/templates/parts/footer.html",
		"cmd/templates/parts/nav.html",
		"cmd/templates/parts/preloader.html",
		"cmd/templates/parts/links.html",
		"cmd/templates/parts/scripts.html",
	)
	if err != nil {
		h.logger.Error("Ошибка вывода главной страницы", zap.Error(err))
		http.Error(w, "Ошибка вывода главной страницы", http.StatusInternalServerError)
		return
	}

	err = h.ExecuteTemplate(t, w, "index", nil)
	if err != nil {
		h.logger.Error("Ошибка вывода главной страницы", zap.Error(err))
		http.Error(w, "Ошибка вывода главной страницы", http.StatusInternalServerError)
		return
	}
}

// AuthView обрабатывает запрос на отображение страницы авторизации.
func (h *Handler) AuthView(w http.ResponseWriter, r *http.Request) {

	t, err := h.ParseTemplate(
		"cmd/templates/auth.html",
		"cmd/templates/parts/preloader.html",
		"cmd/templates/parts/links.html",
		"cmd/templates/parts/scripts.html",
	)
	if err != nil {
		h.logger.Error("Ошибка вывода страницы авторизации", zap.Error(err))
		http.Error(w, "Ошибка вывода страницы авторизации", http.StatusInternalServerError)
		return
	}

	err = h.ExecuteTemplate(t, w, "auth", nil)
	if err != nil {
		h.logger.Error("Ошибка вывода страницы авторизации", zap.Error(err))
		http.Error(w, "Ошибка вывода страницы авторизации", http.StatusInternalServerError)
		return
	}
}

// PuppiesView обрабатывает запрос на отображение страницы с щенками.
func (h *Handler) PuppiesView(w http.ResponseWriter, r *http.Request, archived bool) {
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

	pagedPuppies, puppyReviews, totalPages, err := h.Services.PuppiesGet(
		validatedChocolates, validatedGenders, idPuppy, readyToMove, page, archived,
	)
	if err != nil {
		h.logger.Error("Ошибка при получении данных о щенках", zap.Error(err))
		http.Error(w, "Ошибка при получении данных о щенках", http.StatusInternalServerError)
		return
	}

	getParams := contains(validatedChocolates, validatedGenders, readyToMove)

	if archived {
		t := template.Must(
			template.New("puppyArchive").Funcs(sprig.FuncMap()).ParseFiles(
				"cmd/templates/menu_archive.html",
				"cmd/templates/parts/footer.html",
				"cmd/templates/parts/nav.html",
				"cmd/templates/parts/preloader.html",
				"cmd/templates/parts/links.html",
				"cmd/templates/parts/scripts.html",
			),
		)

		err = t.ExecuteTemplate(
			w, "puppyArchive", struct {
				SelectedColors      []string
				SelectedGenders     []string
				IsReadyToMove       string
				GetParams           string
				Puppies             []domain.Puppy
				PuppiesWithFeedback map[int]int
				TotalPages          int
				CurrentPage         int
			}{
				SelectedColors:      chocolates,
				SelectedGenders:     genders,
				IsReadyToMove:       readyToMove,
				GetParams:           getParams,
				Puppies:             pagedPuppies,
				PuppiesWithFeedback: puppyReviews,
				TotalPages:          totalPages,
				CurrentPage:         page,
			},
		)
		if err != nil {
			h.logger.Error("Ошибка вывода страницы с щенками", zap.Error(err))
		}
	} else {

		t := template.Must(
			template.New("puppyMenu").Funcs(sprig.FuncMap()).ParseFiles(
				"cmd/templates/menu_puppy.html",
				"cmd/templates/parts/footer.html",
				"cmd/templates/parts/nav.html",
				"cmd/templates/parts/preloader.html",
				"cmd/templates/parts/links.html",
				"cmd/templates/parts/scripts.html",
			),
		)

		err = t.ExecuteTemplate(
			w, "puppyMenu", struct {
				SelectedColors  []string
				SelectedGenders []string
				IsReadyToMove   string
				GetParams       string
				Puppies         []domain.Puppy
				TotalPages      int
				CurrentPage     int
			}{
				SelectedColors:  chocolates,
				SelectedGenders: genders,
				IsReadyToMove:   readyToMove,
				GetParams:       getParams,
				Puppies:         pagedPuppies,
				TotalPages:      totalPages,
				CurrentPage:     page,
			},
		)
		if err != nil {
			h.logger.Error("Ошибка вывода страницы с щенками", zap.Error(err))
		}
	}
}

// PuppyView обрабатывает запрос на отображение страницы с щенком.
func (h *Handler) PuppyView(w http.ResponseWriter, r *http.Request) {
	idPuppy := chi.URLParam(r, "id")

	// Проверяем валидность ID
	if !isValidID(idPuppy, 1000) {
		h.logger.Error("Неверный идентификатор щенка", zap.String("id", idPuppy))
		http.Error(w, "Неверный идентификатор щенка", http.StatusNotFound)
		return
	}

	verified := "true"

	puppyInfo, motherInfo, fatherInfo, err := h.Services.PuppyGet(idPuppy)
	if err != nil {
		if err != pgx.ErrNoRows {
			h.logger.Error("Ошибка сервиса: не удалось получить информацию о щенке", zap.Error(err))
			http.Error(w, "Ошибка сервиса: не удалось получить информацию о щенке", http.StatusInternalServerError)
			return
		}
	}
	feedback, err := h.Services.FeedbackGet(idPuppy, verified)
	if err != nil {
		if err != pgx.ErrNoRows {
			h.logger.Error("Ошибка сервиса: не удалось получить информацию об отзыве", zap.Error(err))
			http.Error(w, "Ошибка сервиса: не удалось получить информацию об отзыве", http.StatusInternalServerError)
			return
		}
	}

	t := template.Must(
		template.New("puppyView").Funcs(sprig.FuncMap()).ParseFiles(
			"cmd/templates/puppy.html",
			"cmd/templates/parts/footer.html",
			"cmd/templates/parts/nav.html",
			"cmd/templates/parts/preloader.html",
			"cmd/templates/parts/links.html",
			"cmd/templates/parts/scripts.html",
		),
	)

	err = h.ExecuteTemplate(
		t, w, "puppyView", struct {
			Puppy    *domain.Puppy
			Mother   *domain.Dog
			Father   *domain.Dog
			Feedback *domain.Feedback
		}{
			Puppy:    puppyInfo,
			Mother:   motherInfo,
			Father:   fatherInfo,
			Feedback: feedback,
		},
	)
	if err != nil {
		h.logger.Error("Ошибка вывода страницы с щенками", zap.Error(err))
		http.Error(w, "Ошибка сервера: не удалось отобразить страницу", http.StatusInternalServerError)
		return
	}
}

// ReviewsView обрабатывает запрос на отображение страницы с отзывами.
func (h *Handler) ReviewsView(w http.ResponseWriter, r *http.Request) {
	lastUrlQuery := r.URL.RequestURI
	println(lastUrlQuery)
	// Получаем параметр страницы из URL
	pageStr := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		h.logger.Error("Ошибка при обработке параметра страницы", zap.Error(err))
		page = 1 // если параметр страницы отсутствует или некорректен, установить его в 1
	}

	checked := true
	var idFeedback string
	reviews, puppyNames, err := h.Services.ReviewsGet(idFeedback, checked)
	pagedReviews, totalPages, err := getPagedReviews(reviews, page, 2)
	if err != nil {
		h.logger.Error("Ошибка при получении отзывов", zap.Error(err))
	}

	t := template.Must(
		template.New("reviews").Funcs(sprig.FuncMap()).ParseFiles(
			"cmd/templates/reviews.html",
			"cmd/templates/parts/footer.html",
			"cmd/templates/parts/nav.html",
			"cmd/templates/parts/preloader.html",
			"cmd/templates/parts/links.html",
			"cmd/templates/parts/scripts.html",
		),
	)

	err = t.ExecuteTemplate(
		w, "reviews", struct {
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

// NewFeedbackView обрабатывает запрос на отображение страницы для добавления отзыва.
func (h *Handler) NewFeedbackView(w http.ResponseWriter, r *http.Request) {
	t := template.Must(
		template.New("newFeedback").Funcs(sprig.FuncMap()).ParseFiles(
			"cmd/templates/new_feedback.html",
			"cmd/templates/parts/footer.html",
			"cmd/templates/parts/nav.html",
			"cmd/templates/parts/preloader.html",
			"cmd/templates/parts/links.html",
			"cmd/templates/parts/scripts.html",
		),
	)

	err := h.ExecuteTemplate(
		t, w, "newFeedback", struct {
		}{},
	)
	if err != nil {
		h.logger.Error("Ошибка вывода страницы для добавления отзыва", zap.Error(err))
		http.Error(w, "Ошибка вывода страницы для добавления отзыва", http.StatusInternalServerError)
		return
	}
}

// ContactsView обрабатывает запрос на отображение страницы контактов.
func (h *Handler) ContactsView(w http.ResponseWriter, r *http.Request) {
	t := template.Must(
		template.New("Contacts").Funcs(sprig.FuncMap()).ParseFiles(
			"cmd/templates/contacts.html",
			"cmd/templates/parts/footer.html",
			"cmd/templates/parts/nav.html",
			"cmd/templates/parts/preloader.html",
			"cmd/templates/parts/links.html",
			"cmd/templates/parts/scripts.html",
		),
	)

	err := h.ExecuteTemplate(
		t, w, "Contacts", struct {
		}{},
	)
	if err != nil {
		h.logger.Error("Ошибка вывода страницы контактов", zap.Error(err))
		http.Error(w, "Ошибка вывода страницы контактов", http.StatusInternalServerError)
		return
	}
}

// ArticleView обрабатывает запрос на отображение страницы со статьей.
func (h *Handler) ArticleView(w http.ResponseWriter, r *http.Request) {
	t := template.Must(
		template.New("article").Funcs(sprig.FuncMap()).ParseFiles(
			"cmd/templates/article.html",
			"cmd/templates/parts/footer.html",
			"cmd/templates/parts/nav.html",
			"cmd/templates/parts/preloader.html",
			"cmd/templates/parts/links.html",
			"cmd/templates/parts/scripts.html",
		),
	)

	err := h.ExecuteTemplate(
		t, w, "article", struct {
		}{},
	)
	if err != nil {
		h.logger.Error("Ошибка вывода страницы статьи", zap.Error(err))
		http.Error(w, "Ошибка вывода страницы статьи", http.StatusInternalServerError)
		return
	}
}

// NotFoundView обрабатывает не найденные маршруты.
func (h *Handler) NotFoundView(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	t := template.Must(
		template.New("notFound").Funcs(sprig.FuncMap()).ParseFiles(
			"cmd/templates/not_found.html",
			"cmd/templates/parts/footer.html",
			"cmd/templates/parts/nav.html",
			"cmd/templates/parts/preloader.html",
			"cmd/templates/parts/links.html",
			"cmd/templates/parts/scripts.html",
		),
	)

	err := h.ExecuteTemplate(
		t, w, "notFound", struct {
		}{},
	)
	if err != nil {
		h.logger.Error("Ошибка вывода страницы для кода 404", zap.Error(err))
		http.Error(w, "Ошибка вывода страницы для кода 404", http.StatusInternalServerError)
		return
	}
}
